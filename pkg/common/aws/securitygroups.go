package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/openshift/osde2e/pkg/common/config"
)

// CleanupSecurityGroups deletes all non-default security groups in orphaned osde2e VPCs
// whose CloudFormation stacks are in DELETE_FAILED state. Leftover security groups
// (e.g. from OCPBUGS-74960) block CloudFormation stack deletion, so removing them
// allows the subsequent --vpc cleanup to succeed.
func (CcsAwsSession *ccsAwsSession) CleanupSecurityGroups(activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	// Find all osde2e VPCs
	results, err := CcsAwsSession.ec2.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String("osde2e-*")},
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

	cfnClient := cloudformation.New(CcsAwsSession.session)

	for _, vpc := range results.Vpcs {
		vpcID := aws.StringValue(vpc.VpcId)

		var vpcName string
		for _, tag := range vpc.Tags {
			if aws.StringValue(tag.Key) == "Name" {
				vpcName = aws.StringValue(tag.Value)
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
		stackResp, err := cfnClient.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(vpcStackName),
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ValidationError" {
				// Stack does not exist — nothing to clean up
				continue
			}
			log.Printf("Warning: failed to describe stack %s: %v\n", vpcStackName, err)
			continue
		}

		if len(stackResp.Stacks) == 0 {
			continue
		}
		stackStatus := aws.StringValue(stackResp.Stacks[0].StackStatus)
		if stackStatus != "DELETE_FAILED" {
			log.Printf("Stack %s is in %s state, skipping security group cleanup\n", vpcStackName, stackStatus)
			continue
		}

		// Find all non-default security groups in this VPC
		sgResult, err := CcsAwsSession.ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []*string{aws.String(vpcID)},
				},
			},
		})
		if err != nil {
			log.Printf("Warning: failed to describe security groups for VPC %s: %v\n", vpcID, err)
			continue
		}

		for _, sg := range sgResult.SecurityGroups {
			sgID := aws.StringValue(sg.GroupId)
			sgName := aws.StringValue(sg.GroupName)

			if sgName == "default" {
				continue
			}

			if dryrun {
				log.Printf("Would delete security group %s (%s) in VPC %s\n", sgID, sgName, vpcID)
				continue
			}

			// Revoke all ingress/egress rules before deleting, to avoid DependencyViolation errors
			if len(sg.IpPermissions) > 0 {
				_, err := CcsAwsSession.ec2.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
					GroupId:       aws.String(sgID),
					IpPermissions: sg.IpPermissions,
				})
				if err != nil {
					log.Printf("Warning: failed to revoke ingress rules for SG %s: %v\n", sgID, err)
				}
			}

			if len(sg.IpPermissionsEgress) > 0 {
				_, err := CcsAwsSession.ec2.RevokeSecurityGroupEgress(&ec2.RevokeSecurityGroupEgressInput{
					GroupId:       aws.String(sgID),
					IpPermissions: sg.IpPermissionsEgress,
				})
				if err != nil {
					log.Printf("Warning: failed to revoke egress rules for SG %s: %v\n", sgID, err)
				}
			}

			log.Printf("Deleting security group %s (%s) in VPC %s\n", sgID, sgName, vpcID)
			_, err := CcsAwsSession.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
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
