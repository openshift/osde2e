package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	tagKeyForExemptEC2Instances = "osde2e-proxy"
)

var ErrTerminateEC2Instances = fmt.Errorf("unable to terminate EC2 instances")

// isEC2InstanceFromActiveCluster checks if an EC2 instance belongs to an active cluster
// Returns true if the instance should be skipped (belongs to active cluster), false if it can be cleaned up
func isEC2InstanceFromActiveCluster(instanceName string, activeClusters map[string]bool) bool {
	// Extract cluster name from instance name
	// Example: "osde2e-i5u38-master-0" or "osde2e-i5u38-worker-1" -> "osde2e-i5u38"
	re := regexp.MustCompile(`^(osde2e-[^-]+)-`)
	matches := re.FindStringSubmatch(instanceName)
	if len(matches) >= 2 {
		clusterName := matches[1]
		if activeClusters[clusterName] {
			log.Printf("Skipping EC2 instance for active cluster %s: %s\n", clusterName, instanceName)
			return true
		}
	}
	return false
}

// Hypershift Test Helper Function:
// This function is used to validate the worker nodes displayed by the cluster are the same as the worker nodes displayed by the AWS account.
func (CcsAwsSession *ccsAwsSession) CheckIfEC2ExistBasedOnNodeName(ctx context.Context, nodeName string) (bool, error) {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return false, err
	}

	ec2Instances, err := CcsAwsSession.ec2.DescribeInstances(ctx, &ec2v2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("private-dns-name"),
				Values: []string{nodeName},
			},
		},
	})
	if err != nil {
		return false, err
	}

	return len(ec2Instances.Reservations) > 0, nil
}

// ReleaseElasticIPs releases elastic IPs from loaded aws session. If an instance is
// associated with it, we skip its deletion and log tag name. Dryrun returns aws Error
// from AWS api and is logged.
func (CcsAwsSession *ccsAwsSession) ReleaseElasticIPs(ctx context.Context, dryrun bool, sendSummary bool,
	errorBuilder *strings.Builder,
) (counters Counters, err error) {
	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	results, err := CcsAwsSession.ec2.DescribeAddresses(ctx, &ec2v2.DescribeAddressesInput{})
	if err != nil {
		return counters, err
	}
	fmt.Printf("Addresses found: %d\n", len(results.Addresses))

	for _, address := range results.Addresses {
		if address.AssociationId == nil {
			_, err := CcsAwsSession.ec2.ReleaseAddress(ctx, &ec2v2.ReleaseAddressInput{
				AllocationId: address.AllocationId,
				DryRun:       aws.Bool(dryrun),
			})
			if err == nil {
				counters.Deleted++
				fmt.Printf("Address deleted: %s\n", aws.ToString(address.PublicIp))
			} else {
				counters.Failed++
				errorMsg := fmt.Sprintf("Address %s not deleted: %s\n", aws.ToString(address.PublicIp), err.Error())
				fmt.Println(errorMsg)
				if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
					errorBuilder.WriteString(errorMsg)
				}
			}
		} else {
			if address.NetworkInterfaceId != nil {
				fmt.Printf("Skipping address %s still allocated to network interface id %s \n", aws.ToString(address.PublicIp), aws.ToString(address.NetworkInterfaceId))
			} else {
				fmt.Printf("Skipping address %s (associated but no network interface ID)\n", aws.ToString(address.PublicIp))
			}
		}
	}
	fmt.Printf("Finished elastic IP cleanup. Deleted %d addresses.", counters.Deleted)

	return counters, nil
}

// TerminateEC2Instances finds EC2 instances, then terminates these EC2 instances.
// Ignores EC2 instances with tag Name "osde2e-proxy*" and instances belonging to active clusters.
func (CcsAwsSession *ccsAwsSession) TerminateEC2Instances(ctx context.Context, activeClusters map[string]bool, dryrun bool) (counters Counters, err error) {
	err = CcsAwsSession.GetAWSSessions()
	if err != nil {
		return counters, err
	}

	result, err := CcsAwsSession.ec2.DescribeInstances(ctx, &ec2v2.DescribeInstancesInput{})
	if err != nil {
		return counters, err
	}

	type instanceToDelete struct {
		id   string
		name string
	}

	var instancesToDelete []instanceToDelete
	for _, reservation := range result.Reservations {
		// Each reservation typically has only 1 instance
		instance := reservation.Instances[0]
		for _, tag := range instance.Tags {
			if aws.ToString(tag.Key) != "Name" || strings.Contains(aws.ToString(tag.Value), tagKeyForExemptEC2Instances) || isEC2InstanceFromActiveCluster(aws.ToString(tag.Value), activeClusters) {
				continue
			}
			instancesToDelete = append(instancesToDelete, instanceToDelete{
				id:   aws.ToString(instance.InstanceId),
				name: aws.ToString(tag.Value),
			})
			fmt.Printf("Instance %s (%s) will be deleted\n", aws.ToString(instance.InstanceId), aws.ToString(tag.Value))
			break
		}
	}

	ec2ErrorBuilder := strings.Builder{}
	if !dryrun {
		for _, instance := range instancesToDelete {
			_, err := CcsAwsSession.ec2.TerminateInstances(ctx, &ec2v2.TerminateInstancesInput{
				InstanceIds: []string{instance.id},
			})
			if err != nil {
				errorMessage := fmt.Sprintf("Error terminating instance %s (%s): %s\n", instance.id, instance.name, err.Error())
				ec2ErrorBuilder.WriteString(errorMessage)
				fmt.Print(errorMessage)
				counters.Failed++
			} else {
				counters.Deleted++
				fmt.Printf("Instance %s (%s) deleted\n", instance.id, instance.name)
			}
		}
	}
	if ec2ErrorBuilder.Len() == 0 {
		return counters, nil
	}
	return counters, fmt.Errorf("%w: %s", ErrTerminateEC2Instances, ec2ErrorBuilder.String())
}
