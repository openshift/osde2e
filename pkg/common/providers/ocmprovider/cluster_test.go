package ocmprovider_test

import (
	"fmt"
	"testing"

	"github.com/openshift/osde2e/pkg/common/providers/ocmprovider"
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
