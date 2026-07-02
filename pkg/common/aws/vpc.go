package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnv2 "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// CleanupVPCs deletes VPCs that are not associated with any active osde2e cluster.
func (CcsAwsSession *ccsAwsSession) CleanupVPCs(ctx context.Context, activeClusters map[string]bool, dryrun bool, sendSummary bool, errorBuilder *strings.Builder,
) (counters Counters, err error) {
	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	// Get osde2e VPCs from AWS
	results, err := CcsAwsSession.ec2.DescribeVpcs(ctx, &ec2v2.DescribeVpcsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []string{"osde2e-*"},
			},
		},
	})
	if err != nil {
		return counters, err
	}

	if len(results.Vpcs) == 0 {
		log.Printf("No VPCs found\n")
		return counters, nil
	}

	log.Printf("VPCs found: %d\n", len(results.Vpcs))

	var vpcStacks []string
	for _, vpc := range results.Vpcs {
		var vpcName string
		nameTagFound := false
		for _, tag := range vpc.Tags {
			if tag.Key != nil && tag.Value != nil && *tag.Key == "Name" {
				vpcName = *tag.Value
				nameTagFound = true
				break
			}
		}

		// Skip if no Name tag found
		if !nameTagFound {
			log.Printf("Skipping VPC %s with no Name tag\n", aws.ToString(vpc.VpcId))
			continue
		}

		vpcStacks = append(vpcStacks, getClusterNameFromVPCName(vpcName))
	}

	// Create a map with VPC stack names (cluster-name + "-vpc") from active osde2e clusters
	activeVpcStacks := make(map[string]bool)
	for clusterName := range activeClusters {
		vpcStackName := clusterName + "-vpc"
		activeVpcStacks[vpcStackName] = true
		log.Printf("Cluster %s expects VPC stack: %s\n", clusterName, vpcStackName)
	}

	// Create CloudFormation client early to check stack existence
	cfnClient := cfnv2.NewFromConfig(CcsAwsSession.cfg)

	// Only delete VPC stacks that are not associated with any cluster and actually exist
	var orphanedStacks []string
	for _, vpcStackName := range vpcStacks {
		if !activeVpcStacks[vpcStackName] {
			// Check if the CloudFormation stack actually exists before adding to orphaned list
			_, err := cfnClient.DescribeStacks(ctx, &cfnv2.DescribeStacksInput{
				StackName: aws.String(vpcStackName),
			})
			if err != nil {
				// Stack doesn't exist, skip it
				log.Printf("VPC stack %s does not exist in CloudFormation, skipping\n", vpcStackName)
				continue
			}

			log.Printf("Found orphaned VPC stack: %s\n", vpcStackName)
			orphanedStacks = append(orphanedStacks, vpcStackName)
		} else {
			log.Printf("VPC stack %s has corresponding cluster, skipping\n", vpcStackName)
		}
	}

	log.Printf("Found %d orphaned VPC stacks to delete\n", len(orphanedStacks))

	for _, stackName := range orphanedStacks {
		fmt.Printf("Attempting to delete CloudFormation stack: %s\n", stackName)

		if !dryrun {
			_, err := cfnClient.DeleteStack(ctx, &cfnv2.DeleteStackInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				counters.Failed++
				errorMsg := fmt.Sprintf("Failed to delete CloudFormation stack %s: %v\n", stackName, err)
				fmt.Print(errorMsg)
				if sendSummary && errorBuilder.Len() < 10000 {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			waiter := cfnv2.NewStackDeleteCompleteWaiter(cfnClient)
			err = waiter.Wait(ctx, &cfnv2.DescribeStacksInput{
				StackName: aws.String(stackName),
			}, 30*time.Minute)
			if err != nil {
				counters.Failed++
				errorMsg := fmt.Sprintf("Failed waiting for stack deletion %s: %v\n", stackName, err)
				fmt.Print(errorMsg)
				if sendSummary && errorBuilder.Len() < 10000 {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			counters.Deleted++
			log.Printf("AWS VPC stack %s successfully deleted\n", stackName)
		} else {
			log.Printf("Would delete AWS VPC stack %s\n", stackName)
		}
	}

	return counters, nil
}

var vpcNameRegexp = regexp.MustCompile(`^(osde2e-[^-]+)-[^-]+-vpc$`)

// getClusterNameFromVPCName removes the -yyyyy suffix from VPC names that follow the osde2e-xxxxx-yyyyy-vpc format.
func getClusterNameFromVPCName(vpcName string) string {
	matches := vpcNameRegexp.FindStringSubmatch(vpcName)
	if len(matches) == 2 {
		return matches[1] + "-vpc"
	}
	// If pattern doesn't match, return original name
	return vpcName
}
