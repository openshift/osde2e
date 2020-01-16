package events

import (
	"log"
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
		if !eventsAreEqualWithoutOrder(GetListOfEvents(), test.expectedEvents) {
			t.Errorf("The events did not match the expected events for test: %s.", test.name)
		}
	}
}

func eventsAreEqualWithoutOrder(events1, events2 []string) bool {
	sizeEvents1 := len(events1)
	if sizeEvents1 != len(events2) {
		return false
	}

	for i := 0; i < sizeEvents1; i++ {
		curElement := events1[i]
		foundMatch := false
		for j := 0; j < sizeEvents1; j++ {
			if events2[j] == curElement {
				foundMatch = true
				break
			}
		}

		if !foundMatch {
			return false
		}
	}

	return true
}
