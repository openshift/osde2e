package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
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

func (CcsAwsSession *ccsAwsSession) CleanupOpenIDConnectProviders(activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	input := &iam.ListOpenIDConnectProvidersInput{}
	result, err := CcsAwsSession.iam.ListOpenIDConnectProviders(input)
	if err != nil {
		return err
	}

	for _, provider := range result.OpenIDConnectProviderList {
		input := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: provider.Arn,
		}

		// Get the provider
		result, err := CcsAwsSession.iam.GetOpenIDConnectProvider(input)
		if err != nil {
			return err
		}

		// If provider url contains "cloudfront" or "osde2e-", delete it
		if (strings.Contains(*result.Url, providersubstr) || strings.Contains(*result.Url, rolesubstr)) && !isOIDCProviderFromActiveCluster(*result.Url, activeClusters) {

			fmt.Printf("Provider will be deleted: %s (URL: %s)\n", *provider.Arn, *result.Url)

			if !dryrun {
				input := &iam.DeleteOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: provider.Arn,
				}
				_, err := CcsAwsSession.iam.DeleteOpenIDConnectProvider(input)
				if err != nil {
					*failedCounter++
					errorMsg := fmt.Sprintf("OIDC provider %s not deleted: %s\n", *provider.Arn, err.Error())
					fmt.Println(errorMsg)
					if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
						errorBuilder.WriteString(errorMsg)
					}
					return err
				}
				*deletedCounter++
				fmt.Println("Deleted")
			}
		}
	}

	return nil
}

func (CcsAwsSession *ccsAwsSession) CleanupRoles(activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	input := &iam.ListRolesInput{
		MaxItems: aws.Int64(1000),
	}
	result, err := CcsAwsSession.iam.ListRoles(input)
	if err != nil {
		return err
	}

	for _, role := range result.Roles {
		if strings.Contains(*role.Arn, rolesubstr) && !isRoleFromActiveCluster(*role.Arn, activeClusters) {
			fmt.Printf("Role will be deleted: %s\n", *role.RoleName)

			// Remove Roles from Instance Profiles
			instanceProfilesForRoleInputinput := &iam.ListInstanceProfilesForRoleInput{
				RoleName: role.RoleName,
			}
			instanceProfiles, errInstanceProfiles := CcsAwsSession.iam.ListInstanceProfilesForRole(instanceProfilesForRoleInputinput)

			if errInstanceProfiles != nil {
				return fmt.Errorf("error getting instance profiles for role: %s", errInstanceProfiles)
			}

			for _, instanceProfile := range instanceProfiles.InstanceProfiles {
				fmt.Println("Removing role from instance profile: ", *instanceProfile.InstanceProfileName)

				if !dryrun {
					removeRoleFromInstanceProfileInput := &iam.RemoveRoleFromInstanceProfileInput{
						InstanceProfileName: instanceProfile.InstanceProfileName,
						RoleName:            role.RoleName,
					}
					_, errRemoveRoleFromInstanceProfile := CcsAwsSession.iam.RemoveRoleFromInstanceProfile(removeRoleFromInstanceProfileInput)
					if errRemoveRoleFromInstanceProfile != nil {
						*failedCounter++
						errorMsg := fmt.Sprintf("error removing role from instance profile: %s", errRemoveRoleFromInstanceProfile)
						if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
							errorBuilder.WriteString(errorMsg)
						}
						return fmt.Errorf("%s", errorMsg)
					}
					*deletedCounter++
					fmt.Println("Removed")
				}

			}

			// Delete policy inline to the role
			inlineRolePoliciesInput := &iam.ListRolePoliciesInput{
				RoleName: role.RoleName,
			}
			inlinePolicies, errInlineRolePoliciesInput := CcsAwsSession.iam.ListRolePolicies(inlineRolePoliciesInput)

			if errInlineRolePoliciesInput != nil {
				return errInlineRolePoliciesInput
			}

			for _, policy := range inlinePolicies.PolicyNames {
				fmt.Println("Inline policy will be deleted: ", *policy)

				if !dryrun {
					input := &iam.DeleteRolePolicyInput{
						PolicyName: policy,
						RoleName:   role.RoleName,
					}
					_, errInlinePolicies := CcsAwsSession.iam.DeleteRolePolicy(input)
					if errInlinePolicies != nil {
						return fmt.Errorf("error deleting inline policy: %s", errInlinePolicies)
					}
					fmt.Println("Deleted")
				}
			}

			// Delete policy attached to the role
			attachedRolePoliciesInput := &iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			}
			attachedPolicies, errAttachedRolePoliciesInput := CcsAwsSession.iam.ListAttachedRolePolicies(attachedRolePoliciesInput)
			if errAttachedRolePoliciesInput != nil {
				return errAttachedRolePoliciesInput
			}

			for _, policy := range attachedPolicies.AttachedPolicies {
				fmt.Println("Policy will be detached: ", *policy.PolicyName)

				if !dryrun {
					detachInput := &iam.DetachRolePolicyInput{
						PolicyArn: policy.PolicyArn,
						RoleName:  role.RoleName,
					}
					_, errAttachedPolicies := CcsAwsSession.iam.DetachRolePolicy(detachInput)
					if errAttachedPolicies != nil {
						return errAttachedPolicies
					}
					time.Sleep(2 * time.Second)
					fmt.Println("Detached")
				}
			}

			if !dryrun {
				roleInput := &iam.DeleteRoleInput{
					RoleName: role.RoleName,
				}
				// Delete the role
				_, err = CcsAwsSession.iam.DeleteRole(roleInput)
				if err != nil {
					return err
				}
				fmt.Println("Deleted role")
			}
		}
	}

	return nil
}
