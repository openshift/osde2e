package aws

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnv2 "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2v2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	smithy "github.com/aws/smithy-go"
)

type mockSGEC2 struct {
	describeVpcsFn        func(context.Context, *ec2v2.DescribeVpcsInput, ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error)
	describeSGFn          func(context.Context, *ec2v2.DescribeSecurityGroupsInput, ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error)
	revokeIngressFn       func(context.Context, *ec2v2.RevokeSecurityGroupIngressInput, ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error)
	revokeEgressFn        func(context.Context, *ec2v2.RevokeSecurityGroupEgressInput, ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error)
	deleteSecurityGroupFn func(context.Context, *ec2v2.DeleteSecurityGroupInput, ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error)
}

func (m *mockSGEC2) DescribeVpcs(ctx context.Context, input *ec2v2.DescribeVpcsInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
	return m.describeVpcsFn(ctx, input, optFns...)
}

func (m *mockSGEC2) DescribeSecurityGroups(ctx context.Context, input *ec2v2.DescribeSecurityGroupsInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
	return m.describeSGFn(ctx, input, optFns...)
}

func (m *mockSGEC2) RevokeSecurityGroupIngress(ctx context.Context, input *ec2v2.RevokeSecurityGroupIngressInput, optFns ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error) {
	if m.revokeIngressFn != nil {
		return m.revokeIngressFn(ctx, input, optFns...)
	}
	return &ec2v2.RevokeSecurityGroupIngressOutput{}, nil
}

func (m *mockSGEC2) RevokeSecurityGroupEgress(ctx context.Context, input *ec2v2.RevokeSecurityGroupEgressInput, optFns ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error) {
	if m.revokeEgressFn != nil {
		return m.revokeEgressFn(ctx, input, optFns...)
	}
	return &ec2v2.RevokeSecurityGroupEgressOutput{}, nil
}

func (m *mockSGEC2) DeleteSecurityGroup(ctx context.Context, input *ec2v2.DeleteSecurityGroupInput, optFns ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
	return m.deleteSecurityGroupFn(ctx, input, optFns...)
}

type mockSGCFN struct {
	describeStacksFn func(context.Context, *cfnv2.DescribeStacksInput, ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error)
}

func (m *mockSGCFN) DescribeStacks(ctx context.Context, input *cfnv2.DescribeStacksInput, optFns ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
	return m.describeStacksFn(ctx, input, optFns...)
}

func makeVPC(id, name string) ec2types.Vpc {
	vpc := ec2types.Vpc{
		VpcId: aws.String(id),
	}
	if name != "" {
		vpc.Tags = []ec2types.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
		}
	}
	return vpc
}

func makeSG(id, name string, ingress, egress []ec2types.IpPermission) ec2types.SecurityGroup {
	return ec2types.SecurityGroup{
		GroupId:             aws.String(id),
		GroupName:           aws.String(name),
		IpPermissions:       ingress,
		IpPermissionsEgress: egress,
	}
}

func deleteFailedStack(name string) *cfnv2.DescribeStacksOutput {
	return &cfnv2.DescribeStacksOutput{
		Stacks: []cfntypes.Stack{
			{StackName: aws.String(name), StackStatus: cfntypes.StackStatusDeleteFailed},
		},
	}
}

func TestCleanupSecurityGroups_NoVPCs(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{Vpcs: []ec2types.Vpc{}}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, &mockSGCFN{}, nil, false, false, &deleted, &failed, &errBuilder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 || failed != 0 {
		t.Errorf("expected 0 deleted/failed, got %d/%d", deleted, failed)
	}
}

func TestCleanupSecurityGroups_DescribeVpcsError(t *testing.T) {
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return nil, fmt.Errorf("ec2 api error")
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, &mockSGCFN{}, nil, false, false, &deleted, &failed, &errBuilder)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCleanupSecurityGroups_SkipsVPCWithNoNameTag(t *testing.T) {
	cfnCalled := false
	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "")},
			}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			cfnCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			cfnCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}

	activeClusters := map[string]bool{"osde2e-abcde": true}
	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, activeClusters, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return nil, &smithy.GenericAPIError{Code: "ValidationError", Message: "Stack does not exist"}
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return nil, &smithy.GenericAPIError{Code: "InternalError", Message: "something broke"}
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return &cfnv2.DescribeStacksOutput{
				Stacks: []cfntypes.Stack{
					{StackName: aws.String("osde2e-abcde-vpc"), StackStatus: cfntypes.StackStatusCreateComplete},
				},
			}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			sgCalled = true
			return nil, fmt.Errorf("should not be called")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return &cfnv2.DescribeStacksOutput{Stacks: []cfntypes.Stack{}}, nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-default", "default", nil, nil),
					makeSG("sg-custom", "my-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, input *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.ToString(input.GroupId))
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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

	ingress := []ec2types.IpPermission{{IpProtocol: aws.String("tcp")}}
	egress := []ec2types.IpPermission{{IpProtocol: aws.String("-1")}}

	ec2Mock := &mockSGEC2{
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "custom-sg", ingress, egress),
				},
			}, nil
		},
		revokeIngressFn: func(_ context.Context, input *ec2v2.RevokeSecurityGroupIngressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error) {
			revokedIngressIDs = append(revokedIngressIDs, aws.ToString(input.GroupId))
			return &ec2v2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(_ context.Context, input *ec2v2.RevokeSecurityGroupEgressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error) {
			revokedEgressIDs = append(revokedEgressIDs, aws.ToString(input.GroupId))
			return &ec2v2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, input *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.ToString(input.GroupId))
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "custom-sg",
						[]ec2types.IpPermission{{IpProtocol: aws.String("tcp")}},
						[]ec2types.IpPermission{{IpProtocol: aws.String("-1")}},
					),
				},
			}, nil
		},
		revokeIngressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupIngressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error) {
			revokeCalled = true
			return &ec2v2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupEgressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error) {
			revokeCalled = true
			return &ec2v2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, _ *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			deleteCalled = true
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, true, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "custom-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, _ *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			return nil, fmt.Errorf("DependencyViolation: sg has dependencies")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, true, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "custom-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, _ *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			return nil, fmt.Errorf("delete failed")
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{
					makeVPC("vpc-1", "osde2e-aaaaa-bbbbb-vpc"),
					makeVPC("vpc-2", "osde2e-ccccc-ddddd-vpc"),
				},
			}, nil
		},
		describeSGFn: func(_ context.Context, input *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			describeSGCalls++
			vpcID := input.Filters[0].Values[0]
			if vpcID == "vpc-1" {
				return nil, fmt.Errorf("access denied")
			}
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-2", "custom-sg-2", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, _ *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("any-stack"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "custom-sg",
						[]ec2types.IpPermission{{IpProtocol: aws.String("tcp")}},
						[]ec2types.IpPermission{{IpProtocol: aws.String("-1")}},
					),
				},
			}, nil
		},
		revokeIngressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupIngressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error) {
			return nil, fmt.Errorf("revoke ingress failed")
		},
		revokeEgressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupEgressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error) {
			return nil, fmt.Errorf("revoke egress failed")
		},
		deleteSecurityGroupFn: func(_ context.Context, input *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.ToString(input.GroupId))
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-default", "default", nil, nil),
					makeSG("sg-1", "worker-sg", nil, nil),
					makeSG("sg-2", "master-sg", nil, nil),
					makeSG("sg-3", "lb-sg", nil, nil),
				},
			}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, input *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			deletedIDs = append(deletedIDs, aws.ToString(input.GroupId))
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
		describeVpcsFn: func(_ context.Context, _ *ec2v2.DescribeVpcsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeVpcsOutput, error) {
			return &ec2v2.DescribeVpcsOutput{
				Vpcs: []ec2types.Vpc{makeVPC("vpc-1", "osde2e-abcde-fghij-vpc")},
			}, nil
		},
		describeSGFn: func(_ context.Context, _ *ec2v2.DescribeSecurityGroupsInput, _ ...func(*ec2v2.Options)) (*ec2v2.DescribeSecurityGroupsOutput, error) {
			return &ec2v2.DescribeSecurityGroupsOutput{
				SecurityGroups: []ec2types.SecurityGroup{
					makeSG("sg-1", "empty-sg", nil, nil),
				},
			}, nil
		},
		revokeIngressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupIngressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupIngressOutput, error) {
			revokeIngressCalled = true
			return &ec2v2.RevokeSecurityGroupIngressOutput{}, nil
		},
		revokeEgressFn: func(_ context.Context, _ *ec2v2.RevokeSecurityGroupEgressInput, _ ...func(*ec2v2.Options)) (*ec2v2.RevokeSecurityGroupEgressOutput, error) {
			revokeEgressCalled = true
			return &ec2v2.RevokeSecurityGroupEgressOutput{}, nil
		},
		deleteSecurityGroupFn: func(_ context.Context, _ *ec2v2.DeleteSecurityGroupInput, _ ...func(*ec2v2.Options)) (*ec2v2.DeleteSecurityGroupOutput, error) {
			return &ec2v2.DeleteSecurityGroupOutput{}, nil
		},
	}
	cfnMock := &mockSGCFN{
		describeStacksFn: func(_ context.Context, _ *cfnv2.DescribeStacksInput, _ ...func(*cfnv2.Options)) (*cfnv2.DescribeStacksOutput, error) {
			return deleteFailedStack("osde2e-abcde-vpc"), nil
		},
	}

	deleted, failed := 0, 0
	var errBuilder strings.Builder
	err := cleanupSecurityGroups(context.Background(), ec2Mock, cfnMock, nil, false, false, &deleted, &failed, &errBuilder)
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
