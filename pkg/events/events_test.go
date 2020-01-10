package events

import (
	"log"
	"reflect"
	"testing"
)

func TestGetListOfEvents(t *testing.T) {
	tests := []struct {
		name           string
		events         []EventType
		expectedEvents []string
	}{
		{
			name:           "typical test",
			events:         []EventType{NoHiveLogs, InstallFailed, UpgradeFailed},
			expectedEvents: []string{"NoHiveLogs", "InstallFailed", "UpgradeFailed"},
		},
	}

	for _, test := range tests {
		for _, event := range test.events {
			RecordEvent(event)
		}

		log.Printf("Events:\n%v\n%v", GetListOfEvents(), test.expectedEvents)
		if !reflect.DeepEqual(GetListOfEvents(), test.expectedEvents) {
			t.Errorf("The events did not match the expected events for test: %s.", test.name)
		}
	}
}
