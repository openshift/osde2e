package handlers

import "time"

// Now returns the current time - useful for testing with mocking
func Now() time.Time {
	return time.Now()
}