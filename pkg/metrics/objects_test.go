package metrics

import (
	"fmt"
	"sort"
	"testing"

	"github.com/Masterminds/semver/v3"
)

func TestEventEqual(t *testing.T) {
	tests := []struct {
		name          string
		event1        Event
		event2        Event
		shouldBeEqual bool
	}{
		{
			name: "should be equal",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: true,
		},
		{
			name: "should not be equal install versions",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.1"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal upgrade versions",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal cloud provider",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "something-else",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal environment",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal event",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "another-test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal event",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "test-event",
				ClusterID:      "2345678",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job name",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "test-event",
				ClusterID:      "2345678",
				JobName:        "test-job2",
				JobID:          9999,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job ID",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "test-event",
				ClusterID:      "2345678",
				JobName:        "test-job1",
				JobID:          8888,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal timestamp",
			event1: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Event:          "test-event",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      1,
			},
			event2: Event{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Event:          "test-event",
				ClusterID:      "2345678",
				JobName:        "test-job1",
				JobID:          9999,
				Timestamp:      2,
			},
			shouldBeEqual: false,
		},
	}

	for _, test := range tests {
		if test.event1.Equal(test.event2) != test.shouldBeEqual {
			t.Errorf("test %s failed because event1 and event2's Equal returned %t and should have been %t", test.name, !test.shouldBeEqual, test.shouldBeEqual)
		}
	}
}

func TestEventsSorting(t *testing.T) {
	tests := []struct {
		name   string
		events []Event
	}{
		{
			name: "should be sorted",
			events: []Event{
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Event:          "test-event",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Timestamp:      3,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Event:          "test-event",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Timestamp:      2,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Event:          "test-event",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Timestamp:      1,
				},
			},
		},
	}

	for _, test := range tests {
		events := test.events

		sort.Sort(Events(events))

		var lastTimestamp int64 = 0
		for _, event := range events {
			if event.Timestamp < lastTimestamp {
				t.Errorf("list was not sorted by timestamp as expected for test %s", test.name)
			}

			lastTimestamp = event.Timestamp
		}
	}
}

func TestMetadataEqual(t *testing.T) {
	tests := []struct {
		name          string
		metadata1     Metadata
		metadata2     Metadata
		shouldBeEqual bool
	}{
		{
			name: "should be equal",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: true,
		},
		{
			name: "should not be equal install versions",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.1"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal upgrade versions",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal cloud provider",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "another-test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal environment",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "stage",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal metadata name",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "another-test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal cluster ID",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "2345678",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job name",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job2",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job ID",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job2",
				JobID:          8888,
				Value:          12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal value",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job2",
				JobID:          9999,
				Value:          23456,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal timestamp",
			metadata1: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job1",
				JobID:          9999,
				Value:          12345,
				Timestamp:      1,
			},
			metadata2: Metadata{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				MetadataName:   "test-metadata",
				ClusterID:      "1234567",
				JobName:        "test-job2",
				JobID:          9999,
				Value:          12345,
				Timestamp:      2,
			},
			shouldBeEqual: false,
		},
	}

	for _, test := range tests {
		if test.metadata1.Equal(test.metadata2) != test.shouldBeEqual {
			t.Errorf("test %s failed because metadata1 and metadata2's Equal returned %t and should have been %t", test.name, !test.shouldBeEqual, test.shouldBeEqual)
		}
	}
}

func TestMetadataSorting(t *testing.T) {
	tests := []struct {
		name      string
		metadatas []Metadata
	}{
		{
			name: "should be sorted",
			metadatas: []Metadata{
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      3,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      2,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
			},
		},
	}

	for _, test := range tests {
		metadatas := test.metadatas

		sort.Sort(Metadatas(metadatas))

		var lastTimestamp int64 = 0
		for _, metadata := range metadatas {
			if metadata.Timestamp < lastTimestamp {
				t.Errorf("list was not sorted by timestamp as expected for test %s", test.name)
			}

			lastTimestamp = metadata.Timestamp
		}
	}
}

func TestAddonMetadataEqual(t *testing.T) {
	tests := []struct {
		name           string
		addonMetadata1 AddonMetadata
		addonMetadata2 AddonMetadata
		shouldBeEqual  bool
	}{
		{
			name: "should be equal",
			addonMetadata1: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Install,
			},
			addonMetadata2: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Install,
			},
			shouldBeEqual: true,
		},
		{
			name: "should not be equal metadata",
			addonMetadata1: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Install,
			},
			addonMetadata2: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.3"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Install,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal phase",
			addonMetadata1: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Install,
			},
			addonMetadata2: AddonMetadata{
				Metadata: Metadata{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					MetadataName:   "test-metadata",
					ClusterID:      "1234567",
					JobName:        "test-job1",
					JobID:          9999,
					Value:          12345,
					Timestamp:      1,
				},
				Phase: Upgrade,
			},
			shouldBeEqual: false,
		},
	}

	for _, test := range tests {
		if test.addonMetadata1.Equal(test.addonMetadata2) != test.shouldBeEqual {
			t.Errorf("test %s failed because addonMetadata1 and addonMetadata2's Equal returned %t and should have been %t", test.name, !test.shouldBeEqual, test.shouldBeEqual)
		}
	}
}

func TestAddonMetadataSorting(t *testing.T) {
	tests := []struct {
		name           string
		addonMetadatas []AddonMetadata
	}{
		{
			name: "should be sorted",
			addonMetadatas: []AddonMetadata{
				{
					Metadata: Metadata{
						InstallVersion: semver.MustParse("4.1.0"),
						UpgradeVersion: semver.MustParse("4.1.2"),
						CloudProvider:  "test",
						Environment:    "prod",
						MetadataName:   "test-metadata",
						ClusterID:      "1234567",
						JobName:        "test-job1",
						JobID:          9999,
						Value:          12345,
						Timestamp:      3,
					},
					Phase: Install,
				},
				{
					Metadata: Metadata{
						InstallVersion: semver.MustParse("4.1.0"),
						UpgradeVersion: semver.MustParse("4.1.2"),
						CloudProvider:  "test",
						Environment:    "prod",
						MetadataName:   "test-metadata",
						ClusterID:      "1234567",
						JobName:        "test-job1",
						JobID:          9999,
						Value:          12345,
						Timestamp:      2,
					},
					Phase: Install,
				},
				{
					Metadata: Metadata{
						InstallVersion: semver.MustParse("4.1.0"),
						UpgradeVersion: semver.MustParse("4.1.2"),
						CloudProvider:  "test",
						Environment:    "prod",
						MetadataName:   "test-metadata",
						ClusterID:      "1234567",
						JobName:        "test-job1",
						JobID:          9999,
						Value:          12345,
						Timestamp:      1,
					},
					Phase: Install,
				},
			},
		},
	}

	for _, test := range tests {
		addonMetadatas := test.addonMetadatas

		sort.Sort(AddonMetadatas(addonMetadatas))

		var lastTimestamp int64 = 0
		for _, addonMetadata := range addonMetadatas {
			if addonMetadata.Timestamp < lastTimestamp {
				t.Errorf("list was not sorted by timestamp as expected for test %s", test.name)
			}

			lastTimestamp = addonMetadata.Timestamp
		}
	}
}

func TestJUnitResult(t *testing.T) {
	tests := []struct {
		name          string
		jUnitResult1  JUnitResult
		jUnitResult2  JUnitResult
		shouldBeEqual bool
	}{
		{
			name: "should be equal",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: true,
		},
		{
			name: "should not be equal install versions",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.1"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal upgrade versions",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal cloud provider",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "another-test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal environment",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "stage",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal suite",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "another-test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal test name",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "another-test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal result",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Failed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job name",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job2",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal job ID",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          8888,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal phase",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Upgrade,
				Duration:       12345,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal duration",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       23456,
				Timestamp:      1,
			},
			shouldBeEqual: false,
		},
		{
			name: "should not be equal timestamp",
			jUnitResult1: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.2"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      1,
			},
			jUnitResult2: JUnitResult{
				InstallVersion: semver.MustParse("4.1.0"),
				UpgradeVersion: semver.MustParse("4.1.3"),
				CloudProvider:  "test",
				Environment:    "prod",
				Suite:          "test-suite",
				TestName:       "test-name",
				Result:         Passed,
				JobName:        "test-job1",
				JobID:          9999,
				Phase:          Install,
				Duration:       12345,
				Timestamp:      2,
			},
			shouldBeEqual: false,
		},
	}

	for _, test := range tests {
		if test.jUnitResult1.Equal(test.jUnitResult2) != test.shouldBeEqual {
			t.Errorf("test %s failed because jUnitResult1 and jUnitResult2's Equal returned %t and should have been %t", test.name, !test.shouldBeEqual, test.shouldBeEqual)
		}
	}
}

func TestJUnitResultsSorting(t *testing.T) {
	tests := []struct {
		name         string
		jUnitResults []JUnitResult
	}{
		{
			name: "should be sorted",
			jUnitResults: []JUnitResult{
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Suite:          "test-suite",
					TestName:       "test-name",
					Result:         Passed,
					JobName:        "test-job1",
					JobID:          9999,
					Phase:          Install,
					Duration:       12345,
					Timestamp:      3,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Suite:          "test-suite",
					TestName:       "test-name",
					Result:         Passed,
					JobName:        "test-job1",
					JobID:          9999,
					Phase:          Install,
					Duration:       12345,
					Timestamp:      2,
				},
				{
					InstallVersion: semver.MustParse("4.1.0"),
					UpgradeVersion: semver.MustParse("4.1.2"),
					CloudProvider:  "test",
					Environment:    "prod",
					Suite:          "test-suite",
					TestName:       "test-name",
					Result:         Passed,
					JobName:        "test-job1",
					JobID:          9999,
					Phase:          Install,
					Duration:       12345,
					Timestamp:      1,
				},
			},
		},
	}

	for _, test := range tests {
		jUnitResults := test.jUnitResults

		sort.Sort(JUnitResults(jUnitResults))

		var lastTimestamp int64 = 0
		for _, jUnitResult := range jUnitResults {
			if jUnitResult.Timestamp < lastTimestamp {
				t.Errorf("list was not sorted by timestamp as expected for test %s", test.name)
			}

			lastTimestamp = jUnitResult.Timestamp
		}
	}
}

func TestVersionsEqual(t *testing.T) {
	tests := []struct {
		name          string
		version1      *semver.Version
		version2      *semver.Version
		shouldBeEqual bool
	}{
		{
			name:          "should be equal",
			version1:      semver.MustParse("4.1.0"),
			version2:      semver.MustParse("4.1.0"),
			shouldBeEqual: true,
		},
		{
			name:          "should be equal nil",
			version1:      nil,
			version2:      nil,
			shouldBeEqual: true,
		},
		{
			name:          "should not be equal",
			version1:      semver.MustParse("4.1.0"),
			version2:      semver.MustParse("4.1.2"),
			shouldBeEqual: false,
		},
		{
			name:          "should not be equal nil 1",
			version1:      nil,
			version2:      semver.MustParse("4.1.2"),
			shouldBeEqual: false,
		},
		{
			name:          "should not be equal nil 2",
			version1:      semver.MustParse("4.1.0"),
			version2:      nil,
			shouldBeEqual: false,
		},
	}

	for _, test := range tests {
		fmt.Printf("test %s\n", test.name)
		if versionsEqual(test.version1, test.version2) != test.shouldBeEqual {
			t.Errorf("test %s failed because versionsEqual returned %t and it should have returned %t", test.name, !test.shouldBeEqual, test.shouldBeEqual)
		}
	}
}
