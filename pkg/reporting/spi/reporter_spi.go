package spi

// ReporterSPI is the service provider interface for reporting.
type ReporterSPI interface {
	// Name is the name of the reporting interface.
	Name() string

	// GenerateReport will create the report and return a byte array of the report.
	GenerateReport(reportType string) ([]byte, error)
}
