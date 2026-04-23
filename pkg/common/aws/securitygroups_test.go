package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type mockSGEC2 struct {
	describeVpcsFn        func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error)
	describeSGFn          func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
	revokeIngressFn       func(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error)
	revokeEgressFn        func(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error)
	deleteSecurityGroupFn func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)
}

func (m *mockSGEC2) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	return m.describeVpcsFn(input)
}

func (m *mockSGEC2) DescribeSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return m.describeSGFn(input)
}

func (m *mockSGEC2) RevokeSecurityGroupIngress(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	if m.revokeIngressFn != nil {
		return m.revokeIngressFn(input)
	}
	return &ec2.RevokeSecurityGroupIngressOutput{}, nil
}

func (m *mockSGEC2) RevokeSecurityGroupEgress(input *ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	if m.revokeEgressFn != nil {
		return m.revokeEgressFn(input)
	}
	return &ec2.RevokeSecurityGroupEgressOutput{}, nil
}

func (m *mockSGEC2) DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	return m.deleteSecurityGroupFn(input)
}

type mockSGCFN struct {
	describeStacksFn func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
}

func (m *mockSGCFN) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return m.describeStacksFn(input)
}

func makeVPC(id, name string) *ec2.Vpc {
	vpc := &ec2.Vpc{
		VpcId: aws.String(id),
	}
	if name != "" {
		vpc.Tags = []*ec2.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
		}
	}
	return vpc
}

func makeSG(id, name string, ingress, egress []*ec2.IpPermission) *ec2.SecurityGroup {
	return &ec2.SecurityGroup{
		GroupId:             aws.String(id),
		GroupName:           aws.String(name),
		IpPermissions:       ingress,
		IpPermissionsEgress: egress,
	}
}

func deleteFailedStack(name string) *cloudformation.DescribeStacksOutput {
	return &cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			{StackName: aws.String(name), StackStatus: aws.String("DELETE_FAILED")},
		},
	}
}

func TestCleanupSecurityGroups_NoVPCs(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{Vpcs: []*ec2.Vpc{}}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, &mockSGCFN{}, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 || failed != 0 {
		t.Errorf("expected 0 deleted/failed, got %d/%d", deleted, failed)
	}
}

func TestCleanupSecurityGroups_DescribeVpcsError(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return nil, fmt.Errorf("ec2 api error")
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, &mockSGCFN{}, nil, false, false, &deleted, &failed, &errBuilder)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCleanupSecurityGroups_SkipsVPCWithNoNameTag(t *testing.T) {
	cfnCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "")},
			}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			cfnCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfnCalled {
		t.Error("DescribeStacks should not be called for VPCs with no Name tag")
	}
}

func TestCleanupSecurityGroups_SkipsActiveClusterVPC(t *testing.T) {
	cfnCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			cfnCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}

	activeClusters := map[string]bool{"osde2e-abcde": true}
	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, activeClusters, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfnCalled {
		t.Error("DescribeStacks should not be called for VPCs belonging to active clusters")
	}
}

func TestCleanupSecurityGroups_SkipsWhenStackNotFound(t *testing.T) {
	sgCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return nil, awserr.New("ValidationError", "Stack does not exist", nil)
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sgCalled {
		t.Error("DescribeSecurityGroups should not be called when stack does not exist")
	}
}

func TestCleanupSecurityGroups_SkipsNonValidationErrorFromDescribeStacks(t *testing.T) {
	sgCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return nil, awserr.New("InternalError", "something broke", nil)
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sgCalled {
		t.Error("DescribeSecurityGroups should not be called when DescribeStacks fails")
	}
}

func TestCleanupSecurityGroups_SkipsStackNotInDeleteFailed(t *testing.T) {
	sgCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return &cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					{StackName: aws.String("osde2e-abcde-vpc"), StackStatus: aws.String("CREATE_COMPLETE")},
				},
			}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sgCalled {
		t.Error("DescribeSecurityGroups should not be called when stack is not in DELETE_FAILED state")
	}
}

func TestCleanupSecurityGroups_SkipsEmptyStacksResponse(t *testing.T) {
	sgCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return &cloudformation.DescribeStacksOutput{Stacks: []*cloudformation.Stack{}}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sgCalled {
		t.Error("DescribeSecurityGroups should not be called when DescribeStacks returns no stacks")
	}
}

func TestCleanupSecurityGroups_SkipsDefaultSecurityGroup(t *testing.T) {
	var deletedIDs []string
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-default", "default", nil, nil),
					makeSG("sg-custom", "my-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.StringValue(input.GroupId))
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != "sg-custom" {
		t.Errorf("expected only sg-custom to be deleted, got %v", deletedIDs)
	}
}

func TestCleanupSecurityGroups_DeletesWithRuleRevocation(t *testing.T) {
	var revokedIngressIDs, revokedEgressIDs, deletedIDs []string

	ingress := []*ec2.IpPermission{{IpProtocol: aws.String("tcp")}}
	egress := []*ec2.IpPermission{{IpProtocol: aws.String("-1")}}

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "custom-sg", ingress, egress),
				},
			}, nil
		},
		revokeIngressFn: func(input *ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
			revokedIngressIDs = append(revokedIngressIDs, aws.StringValue(input.GroupId))
			return &ec2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(input *ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
			revokedEgressIDs = append(revokedEgressIDs, aws.StringValue(input.GroupId))
			return &ec2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.StringValue(input.GroupId))
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 || failed != 0 {
		t.Errorf("expected 1 deleted / 0 failed, got %d / %d", deleted, failed)
	}
	if len(revokedIngressIDs) != 1 || revokedIngressIDs[0] != "sg-1" {
		t.Errorf("expected ingress revoked for sg-1, got %v", revokedIngressIDs)
	}
	if len(revokedEgressIDs) != 1 || revokedEgressIDs[0] != "sg-1" {
		t.Errorf("expected egress revoked for sg-1, got %v", revokedEgressIDs)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != "sg-1" {
		t.Errorf("expected sg-1 deleted, got %v", deletedIDs)
	}
}

func TestCleanupSecurityGroups_DryRunSkipsDeletion(t *testing.T) {
	deleteCalled := false
	revokeCalled := false

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "custom-sg",
						[]*ec2.IpPermission{{IpProtocol: aws.String("tcp")}},
						[]*ec2.IpPermission{{IpProtocol: aws.String("-1")}},
					),
				},
			}, nil
		},
		revokeIngressFn: func(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
			revokeCalled = true
			return &ec2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
			revokeCalled = true
			return &ec2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			deleteCalled = true
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, true, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleteCalled {
		t.Error("DeleteSecurityGroup should not be called in dry-run mode")
	}
	if revokeCalled {
		t.Error("Revoke calls should not be made in dry-run mode")
	}
	if deleted != 0 || failed != 0 {
		t.Errorf("expected 0 deleted/failed in dry-run, got %d/%d", deleted, failed)
	}
}

func TestCleanupSecurityGroups_DeleteFailureIncrementsCounter(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "custom-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			return nil, fmt.Errorf("DependencyViolation: sg has dependencies")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, true, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if !strings.Contains(errBuilder.String(), "sg-1") {
		t.Errorf("expected error builder to contain sg-1, got: %s", errBuilder.String())
	}
}

func TestCleanupSecurityGroups_SendSummaryFalseDoesNotWriteErrors(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "custom-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			return nil, fmt.Errorf("delete failed")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if errBuilder.Len() != 0 {
		t.Errorf("expected empty error builder when sendSummary=false, got: %s", errBuilder.String())
	}
}

func TestCleanupSecurityGroups_DescribeSGErrorContinues(t *testing.T) {
	describeSGCalls := 0
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{
					makeVPC("vpc-1", "osde2e-aaaaa-bbbbb-vpc"),
					makeVPC("vpc-2", "osde2e-ccccc-ddddd-vpc"),
				},
			}, nil
		},
		describeSGFn: func(input *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			describeSGCalls++
			vpcID := aws.StringValue(input.Filters[0].Values[0])
			if vpcID == "vpc-1" {
				return nil, fmt.Errorf("access denied")
			}
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-2", "custom-sg-2", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("any-stack"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if describeSGCalls != 2 {
		t.Errorf("expected DescribeSecurityGroups called twice, got %d", describeSGCalls)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted (from second VPC), got %d", deleted)
	}
}

func TestCleanupSecurityGroups_RevokeErrorsDoNotBlockDeletion(t *testing.T) {
	var deletedIDs []string

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "custom-sg",
						[]*ec2.IpPermission{{IpProtocol: aws.String("tcp")}},
						[]*ec2.IpPermission{{IpProtocol: aws.String("-1")}},
					),
				},
			}, nil
		},
		revokeIngressFn: func(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
			return nil, fmt.Errorf("revoke ingress failed")
		},
		revokeEgressFn: func(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
			return nil, fmt.Errorf("revoke egress failed")
		},
		deleteSecurityGroupFn: func(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.StringValue(input.GroupId))
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted despite revoke errors, got %d", deleted)
	}
	if len(deletedIDs) != 1 || deletedIDs[0] != "sg-1" {
		t.Errorf("expected sg-1 deleted, got %v", deletedIDs)
	}
}

func TestCleanupSecurityGroups_MultipleSGsInOneVPC(t *testing.T) {
	var deletedIDs []string

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-default", "default", nil, nil),
					makeSG("sg-1", "worker-sg", nil, nil),
					makeSG("sg-2", "master-sg", nil, nil),
					makeSG("sg-3", "lb-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.StringValue(input.GroupId))
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 3 {
		t.Errorf("expected 3 deleted (all non-default), got %d", deleted)
	}
	if failed != 0 {
		t.Errorf("expected 0 failed, got %d", failed)
	}

	expected := map[string]bool{"sg-1": true, "sg-2": true, "sg-3": true}
	for _, id := range deletedIDs {
		if !expected[id] {
			t.Errorf("unexpected deletion of %s", id)
		}
		delete(expected, id)
	}
	if len(expected) > 0 {
		t.Errorf("expected deletions not seen: %v", expected)
	}
}

func TestCleanupSecurityGroups_SkipsNoRulesRevocation(t *testing.T) {
	revokeIngressCalled := false
	revokeEgressCalled := false

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(*ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
			return &ec2.DescribeVpcsOutput{
				Vpcs: []*ec2.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []*ec2.SecurityGroup{
					makeSG("sg-1", "empty-sg", nil, nil),
				},
			}, nil
		},
		revokeIngressFn: func(*ec2.RevokeSecurityGroupIngressInput) (*ec2.RevokeSecurityGroupIngressOutput, error) {
			revokeIngressCalled = true
			return &ec2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(*ec2.RevokeSecurityGroupEgressInput) (*ec2.RevokeSecurityGroupEgressOutput, error) {
			revokeEgressCalled = true
			return &ec2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
			return &ec2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokeIngressCalled {
		t.Error("RevokeSecurityGroupIngress should not be called when there are no ingress rules")
	}
	if revokeEgressCalled {
		t.Error("RevokeSecurityGroupEgress should not be called when there are no egress rules")
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}
