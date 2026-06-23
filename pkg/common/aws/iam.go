package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamv2 "github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/openshift/osde2e/pkg/common/config"
)

var (
	rolesubstr     = "osde2e"
	providersubstr = "cloudfront"
)

// isOIDCProviderFromActiveCluster checks if an OIDC provider belongs to an active cluster
// Returns true if the provider should be skipped (belongs to active cluster), false if it can be cleaned up
func isOIDCProviderFromActiveCluster(url string, activeClusters map[string]bool) bool {
	// Extract cluster name from OIDC URL
	// Example: "osde2e-i5u38-oidc-t8i8.s3.us-west-2.amazonaws.com" -> "osde2e-i5u38"
	re := regexp.MustCompile(`^(osde2e-[^-]+)-oidc-`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		clusterName := matches[1]
		if activeClusters[clusterName] {
			log.Printf("Skipping OIDC provider for active cluster %s: %s\n", clusterName, url)
			return true
		}
	}
	return false
}

// isRoleFromActiveCluster checks if an IAM role belongs to an active cluster
// Returns true if the role should be skipped (belongs to active cluster), false if it can be cleaned up
func isRoleFromActiveCluster(roleArn string, activeClusters map[string]bool) bool {
	// Extract cluster name from role ARN
	// Example: "arn:aws:iam::123456789012:role/osde2e-i5u38-installer-role" -> "osde2e-i5u38"
	re := regexp.MustCompile(`osde2e-[^-]+-`)
	matches := re.FindStringSubmatch(roleArn)
	if len(matches) >= 1 {
		// Remove the trailing dash to get the cluster name
		clusterName := strings.TrimSuffix(matches[0], "-")
		if activeClusters[clusterName] {
			log.Printf("Skipping IAM role for active cluster %s: %s\n", clusterName, roleArn)
			return true
		}
	}
	return false
}

func (CcsAwsSession *ccsAwsSession) CleanupOpenIDConnectProviders(ctx context.Context, activeClusters map[string]bool, dryrun bool, sendSummary bool,
	errorBuilder *strings.Builder,
) (counters Counters, err error) {
	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	result, err := CcsAwsSession.iam.ListOpenIDConnectProviders(ctx, &iamv2.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return counters, err
	}

	recordOidcFailure := func(arn, detail string) {
		counters.Failed++
		msg := fmt.Sprintf("OIDC provider %s: %s\n", arn, detail)
		fmt.Print(msg)
		if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
			errorBuilder.WriteString(msg)
		}
	}

	for _, provider := range result.OpenIDConnectProviderList {
		arn := aws.ToString(provider.Arn)
		if arn == "" {
			continue
		}

		output, errGet := CcsAwsSession.iam.GetOpenIDConnectProvider(ctx, &iamv2.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: provider.Arn,
		})
		if errGet != nil {
			recordOidcFailure(arn, fmt.Sprintf("get provider: %v", errGet))
			continue
		}
		if output.Url == nil {
			continue
		}
		url := aws.ToString(output.Url)

		// If provider url contains "cloudfront" or "osde2e-", delete it
		if !strings.Contains(url, providersubstr) && !strings.Contains(url, rolesubstr) || isOIDCProviderFromActiveCluster(url, activeClusters) {
			continue
		}

		fmt.Printf("Provider will be deleted: %s (URL: %s)\n", arn, url)

		if !dryrun {
			_, errDel := CcsAwsSession.iam.DeleteOpenIDConnectProvider(ctx, &iamv2.DeleteOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: provider.Arn,
			})
			if errDel != nil {
				recordOidcFailure(arn, fmt.Sprintf("not deleted: %v", errDel))
				continue
			}
			counters.Deleted++
			fmt.Println("Deleted")
		}
	}

	return counters, nil
}

// removeRoleFromAllInstanceProfiles lists instance profiles for the role, then removes
// the role from each. Returns nil on success; on failure the error message is suitable
// for role cleanup reporting (no role name prefix).
func (CcsAwsSession *ccsAwsSession) removeRoleFromAllInstanceProfiles(ctx context.Context, role *iamtypes.Role, dryrun bool) error {
	instanceProfiles, err := CcsAwsSession.iam.ListInstanceProfilesForRole(ctx,
		&iamv2.ListInstanceProfilesForRoleInput{RoleName: role.RoleName},
	)
	if err != nil {
		return fmt.Errorf("list instance profiles: %w", err)
	}

	var errs []string
	for _, instanceProfile := range instanceProfiles.InstanceProfiles {
		if instanceProfile.InstanceProfileName == nil {
			continue
		}
		ipn := aws.ToString(instanceProfile.InstanceProfileName)
		fmt.Println("Removing role from instance profile: ", ipn)
		if !dryrun {
			_, errRm := CcsAwsSession.iam.RemoveRoleFromInstanceProfile(ctx, &iamv2.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: instanceProfile.InstanceProfileName,
				RoleName:            role.RoleName,
			})
			if errRm != nil {
				errs = append(errs, fmt.Sprintf("profile %s: %v", ipn, errRm))
			} else {
				fmt.Println("Removed")
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("instance profile removal: %s", strings.Join(errs, "; "))
	}
	return nil
}

// deleteAllInlineRolePolicies lists and deletes every inline policy on the role.
func (CcsAwsSession *ccsAwsSession) deleteAllInlineRolePolicies(ctx context.Context, role *iamtypes.Role, dryrun bool) error {
	inlinePolicies, err := CcsAwsSession.iam.ListRolePolicies(ctx, &iamv2.ListRolePoliciesInput{
		RoleName: role.RoleName,
	})
	if err != nil {
		return fmt.Errorf("list inline policies: %w", err)
	}

	var errs []string
	for _, policyName := range inlinePolicies.PolicyNames {
		fmt.Println("Inline policy will be deleted: ", policyName)
		if !dryrun {
			_, errDel := CcsAwsSession.iam.DeleteRolePolicy(ctx, &iamv2.DeleteRolePolicyInput{
				PolicyName: aws.String(policyName),
				RoleName:   role.RoleName,
			})
			if errDel != nil {
				errs = append(errs, fmt.Sprintf("policy %s: %v", policyName, errDel))
			} else {
				fmt.Println("Deleted")
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("delete inline policies: %s", strings.Join(errs, "; "))
	}
	return nil
}

// detachAllAttachedRolePolicies lists and detaches every policy attached to the role (managed policies;
// inline policies are removed in deleteAllInlineRolePolicies).
func (CcsAwsSession *ccsAwsSession) detachAllAttachedRolePolicies(ctx context.Context, role *iamtypes.Role, dryrun bool) error {
	attachedPolicies, err := CcsAwsSession.iam.ListAttachedRolePolicies(ctx, &iamv2.ListAttachedRolePoliciesInput{
		RoleName: role.RoleName,
	})
	if err != nil {
		return fmt.Errorf("list attached policies: %w", err)
	}

	var errs []string
	for _, policy := range attachedPolicies.AttachedPolicies {
		if policy.PolicyName == nil || policy.PolicyArn == nil {
			continue
		}
		polName := aws.ToString(policy.PolicyName)
		fmt.Println("Policy will be detached: ", polName)
		if !dryrun {
			_, errDetach := CcsAwsSession.iam.DetachRolePolicy(ctx, &iamv2.DetachRolePolicyInput{
				PolicyArn: policy.PolicyArn,
				RoleName:  role.RoleName,
			})
			if errDetach != nil {
				errs = append(errs, fmt.Sprintf("policy %s: %v", polName, errDetach))
			} else {
				time.Sleep(2 * time.Second)
				fmt.Println("Detached")
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("detach managed policies: %s", strings.Join(errs, "; "))
	}
	return nil
}

// deleteIAMRole calls IAM DeleteRole for the given role.
func (CcsAwsSession *ccsAwsSession) deleteIAMRole(ctx context.Context, role *iamtypes.Role) error {
	_, err := CcsAwsSession.iam.DeleteRole(ctx, &iamv2.DeleteRoleInput{RoleName: role.RoleName})
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

// cleanupOsde2eRole removes one osde2e IAM role: instance profiles, inline and
// attached policies, then DeleteRole. At most one Failed increment per role;
// Deleted increments only after a successful DeleteRole when not dry-run.
func (CcsAwsSession *ccsAwsSession) cleanupOsde2eRole(
	ctx context.Context,
	role *iamtypes.Role,
	roleName string,
	dryrun bool,
	sendSummary bool,
	errorBuilder *strings.Builder,
	counters *Counters,
) {
	recordRoleFailure := func(detail string) {
		counters.Failed++
		msg := fmt.Sprintf("role %s: %s\n", roleName, detail)
		fmt.Print(msg)
		if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
			errorBuilder.WriteString(msg)
		}
	}

	fmt.Printf("Role will be deleted: %s\n", roleName)

	err := CcsAwsSession.removeRoleFromAllInstanceProfiles(ctx, role, dryrun)
	if err != nil {
		recordRoleFailure(err.Error())
		return
	}
	err = CcsAwsSession.deleteAllInlineRolePolicies(ctx, role, dryrun)
	if err != nil {
		recordRoleFailure(err.Error())
		return
	}
	err = CcsAwsSession.detachAllAttachedRolePolicies(ctx, role, dryrun)
	if err != nil {
		recordRoleFailure(err.Error())
		return
	}
	if !dryrun {
		err = CcsAwsSession.deleteIAMRole(ctx, role)
		if err != nil {
			recordRoleFailure(err.Error())
			return
		}
		fmt.Println("Deleted role", roleName)
		counters.Deleted++
	}
}

func (CcsAwsSession *ccsAwsSession) CleanupRoles(ctx context.Context, activeClusters map[string]bool, dryrun bool, sendSummary bool,
	errorBuilder *strings.Builder,
) (counters Counters, err error) {
	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	result, err := CcsAwsSession.iam.ListRoles(ctx, &iamv2.ListRolesInput{
		MaxItems: aws.Int32(1000),
	})
	if err != nil {
		return counters, err
	}

	for i := range result.Roles {
		role := &result.Roles[i]
		if role.RoleName == nil || role.Arn == nil {
			continue
		}
		roleName := aws.ToString(role.RoleName)
		if !strings.Contains(aws.ToString(role.Arn), rolesubstr) || isRoleFromActiveCluster(aws.ToString(role.Arn), activeClusters) {
			continue
		}
		CcsAwsSession.cleanupOsde2eRole(ctx, role, roleName, dryrun, sendSummary, errorBuilder, &counters)
	}

	return counters, nil
}
