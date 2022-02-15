package ocmprovider_test

import (
	"fmt"
	"testing"

	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
	"gotest.tools/v3/assert"
)

type mockCloudRegion struct {
	id      string
	enabled bool
}

func (m mockCloudRegion) ID() string    { return m.id }
func (m mockCloudRegion) Enabled() bool { return m.enabled }

func TestChooseRandomRegion(t *testing.T) {
	type testcase struct {
		Inputs                []ocmprovider.CloudRegion
		ContainsEnabledRegion bool
		Name                  string
	}

	for index, tc := range []testcase{
		{
			Name: "OnlyOneEnabled",
			Inputs: []ocmprovider.CloudRegion{
				mockCloudRegion{
					id:      "foo",
					enabled: true,
				},
			},
			ContainsEnabledRegion: true,
		},
		{
			Name: "HalfEnabled",
			Inputs: []ocmprovider.CloudRegion{
				mockCloudRegion{
					id:      "foo",
					enabled: true,
				},
				mockCloudRegion{
					id:      "bar",
					enabled: false,
				},
			},
			ContainsEnabledRegion: true,
		},
		{
			Name: "OnlyOneDisabled",
			Inputs: []ocmprovider.CloudRegion{
				mockCloudRegion{
					id:      "foo",
					enabled: false,
				},
			},
			ContainsEnabledRegion: false,
		},
	} {
		t.Run(fmt.Sprintf("%d:%s", index, tc.Name), func(t *testing.T) {
			_, found := ocmprovider.ChooseRandomRegion(tc.Inputs...)

			if found != tc.ContainsEnabledRegion {
				t.Errorf("ChooseRandomRegion found=%v, expected %v", found, tc.ContainsEnabledRegion)
			}
		})
	}
}

func TestGetAvailabilityZones(t *testing.T) {
	subnet1, _ := v1.NewSubnetwork().AvailabilityZone("az1").SubnetID("subnet1").Build()
	subnet2, _ := v1.NewSubnetwork().AvailabilityZone("az1").SubnetID("subnet2").Build()
	subnet3, _ := v1.NewSubnetwork().AvailabilityZone("az2").SubnetID("subnet3").Build()
	subnetworksInput := []*v1.Subnetwork{subnet1, subnet2, subnet3}
	configSubnetsInput := []string{"subnet1", "subnet2"}
	result := ocmprovider.GetAvailabilityZones(subnetworksInput, configSubnetsInput)
	assert.DeepEqual(t, result, []string{"az1"})

	configSubnetsInput = []string{"subnet1", "subnet3"}
	result = ocmprovider.GetAvailabilityZones(subnetworksInput, configSubnetsInput)
	assert.DeepEqual(t, result, []string{"az1", "az2"})

	configSubnetsInput = []string{"subnet5"}
	var expected []string
	result = ocmprovider.GetAvailabilityZones(subnetworksInput, configSubnetsInput)
	assert.DeepEqual(t, result, expected)
}
