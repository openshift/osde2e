package db

import (
	"context"
	"fmt"
)

// AlertDataForJob returns a map from the name of a testcase to a slice of occurrences
// of that testcase failing. It will only return recent failures that have happened
// more than once for the same test case.
func (q *Queries) AlertDataForJob(ctx context.Context, jobID int64) (map[string][]ListAlertableRecentTestFailuresRow, error) {
	failures, err := q.ListAlertableFailuresForJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed looking up failures for job %d: %w", jobID, err)
	}
	var names []string
	for _, f := range failures {
		names = append(names, f.Name)
	}
	recentFailures, err := q.ListAlertableRecentTestFailures(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("failed searching for recent failures of %v: %w", names, err)
	}
	namesToInstances := make(map[string][]ListAlertableRecentTestFailuresRow)
	for _, rf := range recentFailures {
		namesToInstances[rf.Name] = append(namesToInstances[rf.Name], rf)
	}
	for name, instances := range namesToInstances {
		if len(instances) < 2 {
			delete(namesToInstances, name)
		}
	}
	return namesToInstances, nil
}
