package spi

import (
	"fmt"
	"sync"
)

type reporterRegistry struct {
	cache map[string]ReporterSPI
	mutex *sync.Mutex
}

var registry = &reporterRegistry{
	cache: map[string]ReporterSPI{},
	mutex: &sync.Mutex{},
}

// RegisterReporter will register a reporter in the list of reporters
func RegisterReporter(reporter ReporterSPI) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	reporterName := reporter.Name()

	if _, ok := registry.cache[reporterName]; ok {
		return fmt.Errorf("reporter %s already exists", reporterName)
	}

	registry.cache[reporterName] = reporter
	return nil
}

// GetReporter will get a named reporter from the reporter cache.
func GetReporter(reporterName string) (ReporterSPI, error) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if reporter, ok := registry.cache[reporterName]; ok {
		if reporter == nil {
			return nil, fmt.Errorf("reporter %s is nil", reporterName)
		}

		return reporter, nil
	}
	return nil, fmt.Errorf("no reporter called %s failed", reporterName)
}

// ListReporters will list all possible reporters.
func ListReporters() []string {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	reporters := []string{}
	for reporter := range registry.cache {
		reporters = append(reporters, reporter)
	}

	return reporters
}
