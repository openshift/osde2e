package aws

// Counters holds per-resource cleanup results for logging and summaries.
type Counters struct {
	Deleted int
	Failed  int
}
