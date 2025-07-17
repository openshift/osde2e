package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (CcsAwsSession *ccsAwsSession) CleanupVPCs(dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

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

	fmt.Printf("VPCs found: %d\n", len(results.Vpcs))

	for _, vpc := range results.Vpcs {
		// Skip default VPCs to avoid accidentally deleting them
		if vpc.IsDefault != nil && *vpc.IsDefault {
			fmt.Printf("Skipping default VPC %s\n", *vpc.VpcId)
			continue
		}

		if !dryrun {
			fmt.Printf("Cleaning up dependencies for VPC %s\n", *vpc.VpcId)

			// Delete dependencies
			if err := CcsAwsSession.deleteVPCDependencies(vpc.VpcId, sendSummary, errorBuilder); err != nil {
				*failedCounter++
				errorMsg := fmt.Sprintf("VPC %s dependency cleanup failed: %s\n", *vpc.VpcId, err.Error())
				fmt.Println(errorMsg)
				if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
					errorBuilder.WriteString(errorMsg)
				}
				continue
			}

			// Delete the VPC itself
			_, err := CcsAwsSession.ec2.DeleteVpc(&ec2.DeleteVpcInput{
				VpcId: vpc.VpcId,
			})
			if err != nil {
				*failedCounter++
				errorMsg := fmt.Sprintf("VPC %s not deleted: %s\n", *vpc.VpcId, err.Error())
				fmt.Println(errorMsg)
				if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
					errorBuilder.WriteString(errorMsg)
				}
			} else {
				*deletedCounter++
				fmt.Printf("VPC deleted: %s\n", *vpc.VpcId)
			}
		} else {
			fmt.Printf("Would delete VPC: %s\n", *vpc.VpcId)
		}
	}

	fmt.Printf("Finished VPC cleanup. Deleted %d VPCs.\n", *deletedCounter)
	return nil
}

// Removes all dependencies from a VPC before deletion
func (CcsAwsSession *ccsAwsSession) deleteVPCDependencies(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	if err := CcsAwsSession.deleteNATGateways(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete NAT gateways: %w", err)
	}

	if err := CcsAwsSession.deleteCustomRoutes(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete custom routes: %w", err)
	}

	if err := CcsAwsSession.deleteNetworkInterfaces(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete network interfaces: %w", err)
	}

	if err := CcsAwsSession.deleteSubnets(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete subnets: %w", err)
	}

	if err := CcsAwsSession.deleteInternetGateways(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete internet gateways: %w", err)
	}

	if err := CcsAwsSession.deleteRouteTables(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete route tables: %w", err)
	}

	if err := CcsAwsSession.deleteSecurityGroups(vpcID, sendSummary, errorBuilder); err != nil {
		return fmt.Errorf("failed to delete security groups: %w", err)
	}

	return nil
}

// Deletes NAT gateways in the VPC
func (CcsAwsSession *ccsAwsSession) deleteNATGateways(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	// Find NAT gateways directly using VPC ID
	natGateways, err := CcsAwsSession.ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	})
	if err != nil {
		return err
	}

	if len(natGateways.NatGateways) == 0 {
		fmt.Printf("No NAT gateways found in VPC %s\n", *vpcID)
		return nil
	}

	// Delete NAT gateways and collect their EIP allocation IDs
	var eipAllocations []*string
	var natGatewayIds []*string
	for _, natGateway := range natGateways.NatGateways {
		fmt.Printf("Deleting NAT Gateway: %s\n", *natGateway.NatGatewayId)

		// Collect EIP allocation IDs for later cleanup
		for _, address := range natGateway.NatGatewayAddresses {
			if address.AllocationId != nil {
				eipAllocations = append(eipAllocations, address.AllocationId)
			}
		}

		_, err := CcsAwsSession.ec2.DeleteNatGateway(&ec2.DeleteNatGatewayInput{
			NatGatewayId: natGateway.NatGatewayId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete NAT Gateway %s: %s\n", *natGateway.NatGatewayId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			natGatewayIds = append(natGatewayIds, natGateway.NatGatewayId)
		}
	}

	// Wait for NAT gateways to be deleted before proceeding
	if len(natGatewayIds) > 0 {
		fmt.Printf("Waiting for %d NAT gateways to be deleted...\n", len(natGatewayIds))

		// Wait up to 5 minutes for NAT gateways to be deleted
		maxWaitTime := 5 * time.Minute
		checkInterval := 30 * time.Second
		startTime := time.Now()

		for time.Since(startTime) < maxWaitTime {
			// Check status of NAT gateways
			natGatewayStatus, err := CcsAwsSession.ec2.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
				NatGatewayIds: natGatewayIds,
			})
			if err != nil {
				break // If we can't describe them, they might be deleted
			}

			allDeleted := true
			for _, natGateway := range natGatewayStatus.NatGateways {
				if natGateway.State != nil && *natGateway.State != "deleted" {
					allDeleted = false
					fmt.Printf("NAT Gateway %s is still in state: %s\n", *natGateway.NatGatewayId, *natGateway.State)
					break
				}
			}

			if allDeleted {
				fmt.Printf("All NAT gateways have been deleted\n")
				break
			}

			fmt.Printf("Still waiting for NAT gateways to be deleted...\n")
			time.Sleep(checkInterval)
		}
	}

	// Release the EIPs that were associated with NAT gateways
	for _, allocationID := range eipAllocations {
		_, err := CcsAwsSession.ec2.ReleaseAddress(&ec2.ReleaseAddressInput{
			AllocationId: allocationID,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to release EIP %s: %s\n", *allocationID, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Released EIP: %s\n", *allocationID)
		}
	}

	return nil
}

// Deletes custom routes from route tables
func (CcsAwsSession *ccsAwsSession) deleteCustomRoutes(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	routeTables, err := CcsAwsSession.ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, routeTable := range routeTables.RouteTables {
		for _, route := range routeTable.Routes {
			// Skip local routes and main route table routes
			if route.GatewayId != nil && *route.GatewayId == "local" {
				continue
			}

			// Skip routes that are already in deleting state
			if route.State != nil && *route.State != "active" {
				continue
			}

			// Delete custom routes
			if route.GatewayId != nil || route.NatGatewayId != nil {
				_, err := CcsAwsSession.ec2.DeleteRoute(&ec2.DeleteRouteInput{
					RouteTableId:         routeTable.RouteTableId,
					DestinationCidrBlock: route.DestinationCidrBlock,
				})
				if err != nil {
					errorMsg := fmt.Sprintf("Failed to delete route %s from route table %s: %s\n",
						aws.StringValue(route.DestinationCidrBlock), *routeTable.RouteTableId, err.Error())
					fmt.Print(errorMsg)
					if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
						errorBuilder.WriteString(errorMsg)
					}
				} else {
					fmt.Printf("Deleted route %s from route table %s\n",
						aws.StringValue(route.DestinationCidrBlock), *routeTable.RouteTableId)
				}
			}
		}
	}

	return nil
}

// Deletes network interfaces in the VPC
func (CcsAwsSession *ccsAwsSession) deleteNetworkInterfaces(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	networkInterfaces, err := CcsAwsSession.ec2.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, networkInterface := range networkInterfaces.NetworkInterfaces {
		// Skip interfaces that are attached to running instances
		if networkInterface.Attachment != nil && networkInterface.Attachment.Status != nil &&
			*networkInterface.Attachment.Status == "attached" {
			fmt.Printf("Skipping attached network interface: %s (attached to %s)\n",
				*networkInterface.NetworkInterfaceId, aws.StringValue(networkInterface.Attachment.InstanceId))
			continue
		}

		// Skip AWS managed ENIs
		if networkInterface.RequesterId != nil && *networkInterface.RequesterId == "amazon-aws" {
			fmt.Printf("Skipping AWS managed ENI: %s (will be cleaned up automatically)\n",
				*networkInterface.NetworkInterfaceId)
			continue
		}

		// Delete user-created network interfaces
		_, err := CcsAwsSession.ec2.DeleteNetworkInterface(&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: networkInterface.NetworkInterfaceId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete network interface %s: %s\n",
				*networkInterface.NetworkInterfaceId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Deleted network interface: %s\n", *networkInterface.NetworkInterfaceId)
		}
	}

	return nil
}

// Deletes subnets in the VPC
func (CcsAwsSession *ccsAwsSession) deleteSubnets(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	subnets, err := CcsAwsSession.ec2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, subnet := range subnets.Subnets {
		_, err := CcsAwsSession.ec2.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete subnet %s: %s\n", *subnet.SubnetId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Deleted subnet: %s\n", *subnet.SubnetId)
		}
	}

	return nil
}

// Detaches and deletes internet gateways from the VPC
func (CcsAwsSession *ccsAwsSession) deleteInternetGateways(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	igws, err := CcsAwsSession.ec2.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, igw := range igws.InternetGateways {
		// Detach from VPC first
		_, err := CcsAwsSession.ec2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
			VpcId:             vpcID,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to detach internet gateway %s: %s\n", *igw.InternetGatewayId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
			continue
		}

		// Then delete the internet gateway
		_, err = CcsAwsSession.ec2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete internet gateway %s: %s\n", *igw.InternetGatewayId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Deleted internet gateway: %s\n", *igw.InternetGatewayId)
		}
	}

	return nil
}

// Deletes custom route tables in the VPC
func (CcsAwsSession *ccsAwsSession) deleteRouteTables(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	routeTables, err := CcsAwsSession.ec2.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, routeTable := range routeTables.RouteTables {
		// Skip the main route table
		isMain := false
		for _, association := range routeTable.Associations {
			if association.Main != nil && *association.Main {
				isMain = true
				break
			}
		}
		if isMain {
			continue
		}

		_, err := CcsAwsSession.ec2.DeleteRouteTable(&ec2.DeleteRouteTableInput{
			RouteTableId: routeTable.RouteTableId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete route table %s: %s\n", *routeTable.RouteTableId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Deleted route table: %s\n", *routeTable.RouteTableId)
		}
	}

	return nil
}

// Deletes custom security groups in the VPC
func (CcsAwsSession *ccsAwsSession) deleteSecurityGroups(vpcID *string, sendSummary bool, errorBuilder *strings.Builder) error {
	securityGroups, err := CcsAwsSession.ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, securityGroup := range securityGroups.SecurityGroups {
		// Skip the default security group
		if securityGroup.GroupName != nil && *securityGroup.GroupName == "default" {
			continue
		}

		_, err := CcsAwsSession.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
			GroupId: securityGroup.GroupId,
		})
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to delete security group %s: %s\n", *securityGroup.GroupId, err.Error())
			fmt.Print(errorMsg)
			if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
				errorBuilder.WriteString(errorMsg)
			}
		} else {
			fmt.Printf("Deleted security group: %s\n", *securityGroup.GroupId)
		}
	}

	return nil
}
