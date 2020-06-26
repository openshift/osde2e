package events

import (
	"fmt"
	"log"
	"testing"

	"github.com/onsi/gomega"
)

func TestEventHandleErrors(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		successEvent  EventType
		failEvent     EventType
		expectedEvent EventType
	}{
		{
			name:          "success",
			err:           nil,
			successEvent:  InstallSuccessful,
			failEvent:     InstallFailed,
			expectedEvent: InstallSuccessful,
		},
		{
			name:          "failure",
			err:           fmt.Errorf("failure"),
			successEvent:  InstallSuccessful,
			failEvent:     InstallFailed,
			expectedEvent: InstallFailed,
		},
	}

	// Make sure the fail handler doesn't panic for these tests
	gomega.RegisterFailHandler(func(message string, callerSkip ...int) {
		// Do nothing
	})
	defer gomega.RegisterFailHandler(nil)

	for _, test := range tests {
		initializeEvents()
		HandleErrorWithEvents(test.err, test.successEvent, test.failEvent)

		events := GetListOfEvents()
		numEvents := len(events)
		if numEvents != 1 {
			t.Errorf("There should only be one event, found %d for test %s.", numEvents, test.name)
		}

		event := EventType(events[0])
		if event != test.expectedEvent {
			t.Errorf("Expected to find event %s, found event %s for test %s.", test.expectedEvent, event, test.name)
		}
	}
}

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
