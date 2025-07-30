package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/openshift/osde2e/pkg/common/spi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// deletes VPCs that are not associated with any active osde2e cluster
func (CcsAwsSession *ccsAwsSession) CleanupVPCs(dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	// Get osde2e VPCs from AWS
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
		log.Printf("No VPCs found\n")
		return nil
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
			log.Printf("Skipping VPC %s with no Name tag\n", *vpc.VpcId)
			continue
		}

		vpcStacks = append(vpcStacks, vpcName)
	}

	provider, err := spi.GetProvider("ocm")
	if err != nil {
		return err
	}

	// Get active osde2e clusters
	clusters, err := provider.ListClusters("properties.MadeByOSDe2e='true'")
	if err != nil {
		return err
	}

	// Create a map with VPC stack name(cluster-name + "-vpc") from active osde2e clusters
	activeVpcStacks := make(map[string]bool)
	for _, cluster := range clusters {
		vpcStackName := cluster.Name() + "-vpc"
		activeVpcStacks[vpcStackName] = true
		log.Printf("Cluster %s expects VPC stack: %s (state: %s)\n", cluster.Name(), vpcStackName, cluster.State())
	}

	// Only delete VPC stacks that are not associated with any cluster
	var orphanedStacks []string
	for _, vpcStackName := range vpcStacks {
		if !activeVpcStacks[vpcStackName] {
			log.Printf("Found orphaned VPC stack: %s\n", vpcStackName)
			orphanedStacks = append(orphanedStacks, vpcStackName)
		} else {
			log.Printf("VPC stack %s has corresponding cluster, skipping\n", vpcStackName)
		}
	}

	// Create CloudFormation client and delete orphaned stacks
	cfnClient := cloudformation.New(CcsAwsSession.session)
	if cfnClient == nil {
		*failedCounter += len(orphanedStacks)
		return fmt.Errorf("failed to create CloudFormation client")
	}

	log.Printf("Found %d orphaned VPC stacks to delete\n", len(orphanedStacks))

	for _, stackName := range orphanedStacks {
		fmt.Printf("Attempting to delete CloudFormation stack: %s\n", stackName)

		if !dryrun {
			_, err := cfnClient.DeleteStack(&cloudformation.DeleteStackInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				*failedCounter++
				errorMsg := fmt.Sprintf("Failed to delete CloudFormation stack %s: %v\n", stackName, err)
				fmt.Print(errorMsg)
				if sendSummary && errorBuilder.Len() < 10000 {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			err = cfnClient.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				*failedCounter++
				errorMsg := fmt.Sprintf("Failed waiting for stack deletion %s: %v\n", stackName, err)
				fmt.Print(errorMsg)
				if sendSummary && errorBuilder.Len() < 10000 {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			*deletedCounter++
			log.Printf("AWS VPC stack %s successfully deleted\n", stackName)
		} else {
			log.Printf("Would delete AWS VPC stack %s\n", stackName)
		}
	}

	return nil
}
