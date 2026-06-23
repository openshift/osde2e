package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnv2 "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	smithy "github.com/aws/smithy-go"
	"github.com/openshift/osde2e/pkg/common/config"
)

type securityGroupEC2 interface {
	DescribeVpcs(ctx context.Context, params *ec2v2.DescribeVpcsInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2v2.DescribeSecurityGroupsInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error)
	RevokeSecurityGroupIngress(ctx context.Context, params *ec2v2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupEgress(ctx context.Context, params *ec2v2.RevokeSecurityGroupEgressInput, optFns ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error)
	DeleteSecurityGroup(ctx context.Context, params *ec2v2.DeleteSecurityGroupInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error)
}

type securityGroupCFN interface {
	DescribeStacks(ctx context.Context, params *cfnv2.DescribeStacksInput, optFns ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error)
}

// CleanupSecurityGroups deletes all non-default security groups in orphaned osde2e VPCs
// whose CloudFormation stacks are in DELETE_FAILED state. Leftover security groups
// (e.g. from OCPBUGS-74960) block CloudFormation stack deletion, so removing them
// allows the subsequent --vpc cleanup to succeed.
func (CcsAwsSession *ccsAwsSession) CleanupSecurityGroups(ctx context.Context, activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	cfnClient := cfnv2.NewFromConfig(CcsAwsSession.cfg)
	return cleanupSecurityGroups(ctx, CcsAwsSession.ec2, cfnClient, activeClusters, dryrun, sendSummary, deletedCounter, failedCounter, errorBuilder)
}

func cleanupSecurityGroups(ctx context.Context, ec2Client securityGroupEC2, cfnClient securityGroupCFN,
	activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	// Find all osde2e VPCs
	results, err := ec2Client.DescribeVpcs(ctx, &ec2v2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{"osde2e-*"},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(results.Vpcs) == 0 {
		log.Println("No osde2e VPCs found for security group cleanup")
		return nil
	}

	log.Printf("Found %d osde2e VPCs to check for leftover security groups\n", len(results.Vpcs))

	// Build active VPC stack names
	activeVpcStacks := make(map[string]bool)
	for clusterName := range activeClusters {
		activeVpcStacks[clusterName+"-vpc"] = true
	}

	for _, vpc := range results.Vpcs {
		vpcID := aws.ToString(vpc.VpcId)

		var vpcName string
		for _, tag := range vpc.Tags {
			if aws.ToString(tag.Key) == "Name" {
				vpcName = aws.ToString(tag.Value)
				break
			}
		}

		if vpcName == "" {
			continue
		}

		vpcStackName := getClusterNameFromVPCName(vpcName)

		// Skip VPCs belonging to active clusters
		if activeVpcStacks[vpcStackName] {
			log.Printf("Skipping security group cleanup for active cluster VPC: %s\n", vpcName)
			continue
		}

		// Verify the stack exists and is in a failed state before cleaning up
		stackResp, err := cfnClient.DescribeStacks(ctx, &cfnv2.DescribeStacksInput{
			StackName: aws.String(vpcStackName),
		})
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ValidationError" {
				// Stack does not exist — nothing to clean up
				continue
			}
			log.Printf("Warning: failed to describe stack %s: %v\n", vpcStackName, err)
			continue
		}

		if len(stackResp.Stacks) == 0 {
			continue
		}
		if stackResp.Stacks[0].StackStatus != cfntypes.StackStatusDeleteFailed {
			log.Printf("Stack %s is in %s state, skipping security group cleanup\n", vpcStackName, stackResp.Stacks[0].StackStatus)
			continue
		}

		// Find all non-default security groups in this VPC
		sgResult, err := ec2Client.DescribeSecurityGroups(ctx, &ec2v2.DescribeSecurityGroupsInput{
			Filters: []ec2types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{vpcID},
				},
			},
		})
		if err != nil {
			log.Printf("Warning: failed to describe security groups for VPC %s: %v\n", vpcID, err)
			continue
		}

		for _, sg := range sgResult.SecurityGroups {
			sgID := aws.ToString(sg.GroupId)
			sgName := aws.ToString(sg.GroupName)

			if sgName == "default" {
				continue
			}

			if dryrun {
				log.Printf("Would delete security group %s (%s) in VPC %s\n", sgID, sgName, vpcID)
				continue
			}

			// Revoke all ingress/egress rules before deleting, to avoid DependencyViolation errors
			if len(sg.IpPermissions) > 0 {
				_, err := ec2Client.RevokeSecurityGroupIngress(ctx, &ec2v2.RevokeSecurityGroupIngressInput{
					GroupId:       aws.String(sgID),
					IpPermissions: sg.IpPermissions,
				})
				if err != nil {
					log.Printf("Warning: failed to revoke ingress rules for SG %s: %v\n", sgID, err)
				}
			}

			if len(sg.IpPermissionsEgress) > 0 {
				_, err := ec2Client.RevokeSecurityGroupEgress(ctx, &ec2v2.RevokeSecurityGroupEgressInput{
					GroupId:       aws.String(sgID),
					IpPermissions: sg.IpPermissionsEgress,
				})
				if err != nil {
					log.Printf("Warning: failed to revoke egress rules for SG %s: %v\n", sgID, err)
				}
			}

			log.Printf("Deleting security group %s (%s) in VPC %s\n", sgID, sgName, vpcID)
			_, err = ec2Client.DeleteSecurityGroup(ctx, &ec2v2.DeleteSecurityGroupInput{
				GroupId: aws.String(sgID),
			})
			if err != nil {
				*failedCounter++
				errorMsg := fmt.Sprintf("Failed to delete security group %s (%s): %v\n", sgID, sgName, err)
				log.Print(errorMsg)
				if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			*deletedCounter++
			log.Printf("Deleted security group %s (%s)\n", sgID, sgName)
		}
	}

	return nil
}
