package aws

import (
	"fmt"
	"strings"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
