package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	tagKeyForExemptEC2Instances = "osde2e-proxy"
)

var ErrTerminateEC2Instances = fmt.Errorf("unable to terminate EC2 instances")

// Hypershift Test Helper Function:
// This function is used to validate the worker nodes displayed by the cluster are the same as the worker nodes displayed by the AWS account.
func (CcsAwsSession *ccsAwsSession) CheckIfEC2ExistBasedOnNodeName(nodeName string) (bool, error) {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return false, err
	}

	ec2Instances, err := CcsAwsSession.ec2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("private-dns-name"),
				Values: []*string{aws.String(nodeName)},
			},
		},
	})
	if err != nil {
		return false, err
	}

	if len(ec2Instances.Reservations) > 0 {
		return true, nil
	}

	return false, nil
}

// ReleaseElasticIPs releases elastic IPs from loaded aws session. If an instance is
// associated with it, we skip its deletion and log tag name. Dryrun returns aws Error
// from AWS api and is logged.
func (CcsAwsSession *ccsAwsSession) ReleaseElasticIPs(dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	results, err := CcsAwsSession.ec2.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err != nil {
		return err
	}
	fmt.Printf("Addresses found: %d\n", len(results.Addresses))

	for _, address := range results.Addresses {
		if address.AssociationId == nil {
			_, err := CcsAwsSession.ec2.ReleaseAddress(&ec2.ReleaseAddressInput{
				AllocationId: address.AllocationId,
				DryRun:       &dryrun,
			})
			if err == nil {
				*deletedCounter++
				fmt.Printf("Address deleted: %s\n", *address.PublicIp)
			} else {
				*failedCounter++
				errorMsg := fmt.Sprintf("Address %s not deleted: %s\n", *address.PublicIp, err.Error())
				fmt.Println(errorMsg)
				if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
					errorBuilder.WriteString(errorMsg)
				}
			}
		} else {
			fmt.Printf("Skipping address %s still allocated to network interface id %s \n", *address.PublicIp, *address.NetworkInterfaceId)
		}
	}
	fmt.Printf("Finished elastic IP cleanup. Deleted %d addresses.", *deletedCounter)

	return nil
}

// TerminateEC2Instances finds EC2 instances older than given duration, then terminates these EC2 instances.
// Ignores EC2 instances with tag Name "osde2e-proxy*".
func (CcsAwsSession *ccsAwsSession) TerminateEC2Instances(olderthan time.Duration, dryrun bool) (int, int, error) {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return 0, 0, err
	}
	result, err := CcsAwsSession.ec2.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return 0, 0, err
	}
	var instanceIds []string
	for _, reservation := range result.Reservations {
		// Each reservation typically has only 1 instance
		instance := reservation.Instances[0]
		if time.Since(*instance.LaunchTime) < olderthan {
			continue
		}
		for _, tag := range instance.Tags {
			if *tag.Key != "Name" || strings.Contains(*tag.Value, tagKeyForExemptEC2Instances) {
				continue
			}
			instanceIds = append(instanceIds, *instance.InstanceId)
			break
		}
	}

	ec2ErrorBuilder := strings.Builder{}
	instancesDeleted := 0
	instancesFailedToDelete := 0
	if !dryrun {
		for _, instanceId := range instanceIds {
			input := &ec2.TerminateInstancesInput{
				InstanceIds: aws.StringSlice([]string{instanceId}),
			}
			_, err := CcsAwsSession.ec2.TerminateInstances(input)
			if err != nil {
				errorMessage := fmt.Sprintf("Error terminating instance %s: %s\n", instanceId, err.Error())
				ec2ErrorBuilder.WriteString(errorMessage)
				fmt.Printf(errorMessage)
				instancesFailedToDelete++
			} else {
				instancesDeleted++
			}
		}
	}
	if ec2ErrorBuilder.Len() == 0 {
		return instancesDeleted, instancesFailedToDelete, nil
	}
	ec2Error := fmt.Errorf("%w: %s", ErrTerminateEC2Instances, ec2ErrorBuilder.String())
	return instancesDeleted, instancesFailedToDelete, ec2Error
}
