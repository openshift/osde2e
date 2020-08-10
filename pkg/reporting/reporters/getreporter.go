package reporters

import (
	"github.com/openshift/osde2e/pkg/reporting/spi"
)

// GetReporter will get a named reporter from the reporter cache.
func GetReporter(reporterName string) (spi.ReporterSPI, error) {
	return spi.GetReporter(reporterName)
}
