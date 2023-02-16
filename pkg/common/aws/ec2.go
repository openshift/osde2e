package aws

import (
	"fmt"

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

// getSecurityGroups gets all security groups based on the security groups input struct
func (CcsAwsSession *ccsAwsSession) getSecurityGroups(securityGroupsInput ec2.DescribeSecurityGroupsInput) ([]string, error) {
	var securityGroupIds []string

	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return nil, err
	}

	securityGroups, err := CcsAwsSession.ec2.DescribeSecurityGroups(&securityGroupsInput)
	if err != nil {
		return nil, fmt.Errorf("error attempting to fetch security groups: %w", err)
	}

	if len(securityGroups.SecurityGroups) == 0 {
		return nil, fmt.Errorf("no security groups found")
	}

	for _, securityGroup := range securityGroups.SecurityGroups {
		securityGroupIds = append(securityGroupIds, *securityGroup.GroupId)
		
	}

	return securityGroupIds, nil
}

// DeleteHyperShiftELBSecurityGroup is a temporary solution to remove elb security group
// created when HyperShift clusters are deleted. Bug: HOSTEDCP-656
func (CcsAwsSession *ccsAwsSession) DeleteHyperShiftELBSecurityGroup(clusterID string) error {
	var securityGroupId string
	var defaultSecurityGroupId string
	var securityGroupRuleId *string

	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	securityGroups, err := CcsAwsSession.getSecurityGroups(ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag-key"),
				Values: []*string{aws.String(fmt.Sprintf("kubernetes.io/cluster/%s", clusterID))},
			},
		},
	})
	if err != nil {
		return err
	}

	// Only will be one security group entry
	securityGroupId = securityGroups[0]

	defaultSecurityGroups, err := CcsAwsSession.getSecurityGroups(ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("ip-permission.group-id"),
				Values: []*string{aws.String(securityGroupId)},
			},
		},
	})
	if err != nil {
		return err
	}

	// Only will be one security group entry
	defaultSecurityGroupId = defaultSecurityGroups[0]

	securityGroupRules, err := CcsAwsSession.ec2.DescribeSecurityGroupRules(&ec2.DescribeSecurityGroupRulesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []*string{aws.String(defaultSecurityGroupId)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error getting security group: '%s' rules: %w", defaultSecurityGroupId, err)
	}
	if len(securityGroupRules.SecurityGroupRules) == 0 {
		return fmt.Errorf("no rules found for security group: '%s'", defaultSecurityGroupId)
	}

	for _, securityGroupRule := range securityGroupRules.SecurityGroupRules {
		if securityGroupRule.ReferencedGroupInfo == nil {
			continue
		}

		if *securityGroupRule.ReferencedGroupInfo.GroupId == securityGroupId {
			securityGroupRuleId = securityGroupRule.SecurityGroupRuleId
			break
		}
	}

	_, err = CcsAwsSession.ec2.RevokeSecurityGroupIngress(&ec2.RevokeSecurityGroupIngressInput{
		GroupId:              &defaultSecurityGroupId,
		SecurityGroupRuleIds: []*string{securityGroupRuleId},
	})
	if err != nil {
		return fmt.Errorf("error revoking security group ingress rule: %w", err)
	}

	_, err = CcsAwsSession.ec2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: &securityGroupId})
	if err != nil {
		return fmt.Errorf("error deleting security group '%s': %w", securityGroupId, err)
	}

	return nil
}
