package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

var (
	rolesubstr     = "osde2e"
	providersubstr = "cloudfront"
)

func (CcsAwsSession *ccsAwsSession) CleanupOpenIDConnectProviders(olderthan time.Duration, dryrun bool) error {
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

		// If provider URL contains "cloudfront" or "osde2e", and is older than given days, delete it
		if (strings.Contains(*result.Url, providersubstr) || strings.Contains(*result.Url, rolesubstr)) && time.Since(*result.CreateDate) > olderthan {
			fmt.Printf("Provider will be deleted: %s\n", *provider.Arn)

			if !dryrun {
				input := &iam.DeleteOpenIDConnectProviderInput{
					OpenIDConnectProviderArn: provider.Arn,
				}
				_, err := CcsAwsSession.iam.DeleteOpenIDConnectProvider(input)
				if err != nil {
					return err
				}
				fmt.Println("Deleted")
			}
		}
	}

	return nil
}

func (CcsAwsSession *ccsAwsSession) CleanupRoles(olderthan time.Duration, dryrun bool) error {
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
		// do not delete OrganizationAccountAccessRole. It's created one time during attachment to parent org.
		if strings.Contains(*role.Arn, rolesubstr) && time.Since(*role.CreateDate) > olderthan && *role.RoleName != "OrganizationAccountAccessRole" {
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
						return fmt.Errorf("error removing role from instance profile: %s", errRemoveRoleFromInstanceProfile)
					}
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

func (CcsAwsSession *ccsAwsSession) CleanupPolicies(olderthan time.Duration, dryrun bool) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}
	input := &iam.ListPoliciesInput{
		MaxItems: aws.Int64(1000),
	}
	result, err := CcsAwsSession.iam.ListPolicies(input)
	if err != nil {
		return err
	}

	for _, policy := range result.Policies {
		if strings.Contains(*policy.Arn, rolesubstr) && time.Since(*policy.CreateDate) > olderthan {
			fmt.Printf("Policy will be deleted: %s", *policy.PolicyName)

			if !dryrun {
				input := &iam.DeletePolicyInput{
					PolicyArn: policy.Arn,
				}
				// Delete the policy
				_, err := CcsAwsSession.iam.DeletePolicy(input)
				if err != nil {
					return err
				}
				fmt.Println("Deleted")
			}
		}
	}

	return nil
}
