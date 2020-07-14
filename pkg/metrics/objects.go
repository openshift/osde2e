package metrics

import (
	"regexp"
	"time"

	"github.com/Masterminds/semver"
)

// AllJobs represents a regex that will collect results from all jobs.
var AllJobs = regexp.MustCompile(".*")

// Phase is a phase of an osde2e run.
type Phase string

// Result is the result of a JUnit test.
type Result string

const (
	// Install phase represents tests that were run after the initial installation of the cluster.
	Install Phase = "install"

	// Upgrade phase represents tests that were run after the upgrade of the cluster.
	Upgrade Phase = "upgrade"

	// UnknownPhase represents tests that were run in a phase that is currently unknown to the metrics library.
	UnknownPhase Phase = "unknown"

	// Passed result represents a JUnitResult that passed acceptably.
	Passed Result = "passed"

	// Failed result represents a JUnitResult that failed.
	Failed Result = "failed"

	// Skipped result represents a JUnitResult that was skipped during a run.
	Skipped Result = "skipped"

	// UnknownResult represents a JUnitResult that is currently unknown to the metrics library.
	UnknownResult Result = "unknown"
)

// Event objects that are recorded by osde2e runs. These typically represent occurrences that are of
// some note. For example, cluste provisioning failure, failure to collect Hive logs, etc.
type Event struct {
	// InstallVersion is the starting install version of the cluster that generated this event.
	InstallVersion *semver.Version

	// UpgradeVersion is the upgrade version of the cluster that generated this event. This can be nil.
	UpgradeVersion *semver.Version

	// CloudProvider is the cluster cloud provider that was used when this event was generated.
	CloudProvider string

	// Environment is the environment that the cluster provider was using during the generation of this event.
	Environment string

	// Event is the name of the event that was recorded.
	Event string

	// ClusterID is the cluster ID of the cluster that was provisioned while generating this event.
	ClusterID string

	// JobName is the name of the job that generated this event.
	JobName string

	// JobID is the job ID number that corresponds to the job that generated this event.
	JobID int64

	// Timestamp is the time when this event was recorded.
	Timestamp int64
}

// Equal will return true if two event objects are equal.
func (e Event) Equal(that Event) bool {
	if !versionsEqual(e.InstallVersion, that.InstallVersion) {
		return false
	}

	if !versionsEqual(e.UpgradeVersion, that.UpgradeVersion) {
		return false
	}

	if e.CloudProvider != that.CloudProvider {
		return false
	}

	if e.Environment != that.Environment {
		return false
	}

	if e.Event != that.Event {
		return false
	}

	if e.ClusterID != that.ClusterID {
		return false
	}

	if e.JobName != that.JobName {
		return false
	}

	if e.JobID != that.JobID {
		return false
	}

	if e.Timestamp != that.Timestamp {
		return false
	}

	return true
}

// Events is a list of events.
type Events []Event

func (e Events) Len() int {
	return len(e)
}

func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Events) Less(i, k int) bool {
	return e[i].Timestamp < e[k].Timestamp
}

// Metadata objects are numerical values associated with metadata calculated by osde2e.
type Metadata struct {
	// InstallVersion is the starting install version of the cluster that generated this metadata.
	InstallVersion *semver.Version

	// UpgradeVersion is the upgrade version of the cluster that generated this metadata. This can be nil.
	UpgradeVersion *semver.Version

	// CloudProvider is the cluster cloud provider that was used when this metadata was generated.
	CloudProvider string

	// Environment is the environment that the cluster provider was using during the generation of this metadata.
	Environment string

	// MetadataName is the name of the metadata that was recorded.
	MetadataName string

	// ClusterID is the cluster ID of the cluster that was provisioned while generating this metadata.
	ClusterID string

	// JobName is the name of the job that generated this metadata.
	JobName string

	// JobID is the job ID number that corresponds to the job that generated this metadata.
	JobID int64

	// Value is the numerical value associated with this metadata.
	Value float64

	// Time is the time when this metadata was recorded.
	Timestamp int64
}

// Equal will return true if two metadata objects are equal.
func (m Metadata) Equal(that Metadata) bool {
	if !versionsEqual(m.InstallVersion, that.InstallVersion) {
		return false
	}

	if !versionsEqual(m.UpgradeVersion, that.UpgradeVersion) {
		return false
	}

	if m.CloudProvider != that.CloudProvider {
		return false
	}

	if m.Environment != that.Environment {
		return false
	}

	if m.MetadataName != that.MetadataName {
		return false
	}

	if m.ClusterID != that.ClusterID {
		return false
	}

	if m.JobName != that.JobName {
		return false
	}

	if m.JobID != that.JobID {
		return false
	}

	if m.Value != that.Value {
		return false
	}

	if m.Timestamp != that.Timestamp {
		return false
	}

	return true
}

// Metadatas is a list of metadata objects.
type Metadatas []Metadata

func (m Metadatas) Len() int {
	return len(m)
}

func (m Metadatas) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Metadatas) Less(i, k int) bool {
	return m[i].Timestamp < m[k].Timestamp
}

// AddonMetadata is numerical data captured by osde2e runs, similar to Metadata. However, this is customizable and
// focused on addon testing.
type AddonMetadata struct {
	Metadata

	// Phase is the test phase where this this metadata was generated in.
	Phase Phase
}

// Equal will return true if two addon metadata objects are equal.
func (a AddonMetadata) Equal(that AddonMetadata) bool {
	if !a.Metadata.Equal(that.Metadata) {
		return false
	}

	if a.Phase != that.Phase {
		return false
	}

	return true
}

// AddonMetadatas is a list of addon metadata objects.
type AddonMetadatas []AddonMetadata

func (a AddonMetadatas) Len() int {
	return len(a)
}

func (a AddonMetadatas) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AddonMetadatas) Less(i, k int) bool {
	return a[i].Timestamp < a[k].Timestamp
}

// JUnitResult represents an individual test that was run over the course of an osde2e run.
type JUnitResult struct {
	// InstallVersion is the starting install version of the cluster that generated this result.
	InstallVersion *semver.Version

	// UpgradeVersion is the upgrade version of the cluster that generated this result. This can be nil.
	UpgradeVersion *semver.Version

	// CloudProvider is the cluster cloud provider that was used when this result was generated.
	CloudProvider string

	// Environment is the environment that the cluster provider was using during the generation of this result.
	Environment string

	// Suite is the name of the test suite that this test belongs to.
	Suite string

	// TestName is the name of the test that was run.
	TestName string

	// Result is the result of this test.
	Result Result

	// ClusterID is the cluster ID of the cluster that was provisioned while generating this result.
	ClusterID string

	// JobName is the name of the job that generated this result.
	JobName string

	// JobID is the job ID number that corresponds to the job that generated this result.
	JobID int64

	// Phase is the test phase where this this result was generated in.
	Phase Phase

	// Duration is the length of time that this test took to run.
	Duration time.Duration

	// Timestamp is the timestamp when this result was recorded.
	Timestamp int64
}

// Equal will return true if two JUnitResult objects are equal.
func (j JUnitResult) Equal(that JUnitResult) bool {
	if !versionsEqual(j.InstallVersion, that.InstallVersion) {
		return false
	}

	if !versionsEqual(j.UpgradeVersion, that.UpgradeVersion) {
		return false
	}

	if j.CloudProvider != that.CloudProvider {
		return false
	}

	if j.Environment != that.Environment {
		return false
	}

	if j.Suite != that.Suite {
		return false
	}

	if j.TestName != that.TestName {
		return false
	}

	if j.Result != that.Result {
		return false
	}

	if j.ClusterID != that.ClusterID {
		return false
	}

	if j.JobName != that.JobName {
		return false
	}

	if j.JobID != that.JobID {
		return false
	}

	if j.Phase != that.Phase {
		return false
	}

	if j.Duration != that.Duration {
		return false
	}

	if j.Timestamp != that.Timestamp {
		return false
	}

	return true
}

// JUnitResults is a list of JUnitResults.
type JUnitResults []JUnitResult

func (jr JUnitResults) Len() int {
	return len(jr)
}

func (jr JUnitResults) Swap(i, j int) {
	jr[i], jr[j] = jr[j], jr[i]
}

func (jr JUnitResults) Less(i, k int) bool {
	return jr[i].Timestamp < jr[k].Timestamp
}

// nil safe semver equivalency
func versionsEqual(version1, version2 *semver.Version) bool {
	return (version1 == nil && version1 == version2) || (version1 != nil && version2 != nil && version1.Equal(version2))
}
