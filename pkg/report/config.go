package report

// DefaultConfig has initial values which are reasonable for most failure reports.
var DefaultConfig = &Config{
	Tests:      DefaultTests,
	DateLayout: DefaultDateLayout,
}

const (
	// DefaultDateLayout is the time layout used to print dates.
	DefaultDateLayout = "January 2, 2006"
)

// DefaultTests are included in reports.
var DefaultTests = []string{
	"BeforeSuite",
	"AfterSuite",
}

// Config contains options for the generation of failure reports.
type Config struct {
	// Envs are the environments that are reported.
	Envs []EnvConfig

	// Jobs that are reported.
	Jobs []JobConfig

	// Tests are the names of tests that are included in the report.
	Tests []string

	// DateLayout defines the format of dates within the report.
	DateLayout string
}

// EnvConfig for environment being reported.
type EnvConfig struct {
	// Name is based on prefix used in GCS.
	Name string

	// SkipJobs are not shown for an env. If equal to []string{"*"} then all jobs are ignored.
	SkipJobs []string
}

// JobConfig for job being reported.
type JobConfig struct {
	// Name of job being run.
	Name string

	// Version being tested.
	Version string
}

// SkipJob is true when a job should not be shown in a report.
func (e EnvConfig) SkipJob(jobName string) bool {
	switch {
	case len(e.SkipJobs) == 0:
		return false
	case len(e.SkipJobs) == 1 && e.SkipJobs[0] == "*":
		return true
	default:
		for _, skipJob := range e.SkipJobs {
			if skipJob == jobName {
				return true
			}
		}
	}
	return false
}
