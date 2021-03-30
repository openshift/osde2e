package events

import (
	"log"
)

// Events records individual events that occur during the execution of osde2e
type Events struct {
	Events map[string]bool
}

// Instance is the global Events instance
var Instance *Events

func init() {
	initializeEvents()
}

func initializeEvents() {
	Instance = &Events{}
	Instance.Events = map[string]bool{}
}

// HandleErrorWithEvents records events depending on the error state.
func HandleErrorWithEvents(err error, successEvent EventType, failEvent EventType) {
	if err != nil {
		log.Printf("Fail event: %v", failEvent)
		RecordEvent(failEvent)
	} else {
		log.Printf("Success event: %v", successEvent)
		RecordEvent(successEvent)
	}
}

// RecordEvent records the given event in the global events instance
func RecordEvent(event EventType) {
	Instance.Events[string(event)] = true
}

// GetListOfEvents gets the list of events that were registered with the event recorder
func GetListOfEvents() []string {
	events := make([]string, len(Instance.Events))

	i := 0
	for k := range Instance.Events {
		events[i] = k
		i++
	}

	return events
}
