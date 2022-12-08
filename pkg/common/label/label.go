// Package label implements a standard set of Ginkgo labels for classifying
// OpenShift Dedicated end to end platform/component tests.
//
// A label is for classifying or grouping a test based on what it is doing,
// what is supports, and it's level of importance in decision making. These
// labels can be used to target specific subsets of the overall suite during
// execution by providing the Ginkgo/osde2e `--label-filter` flag.
//
// Within the `Describe` or `It` of a test, one to many labels can be included
// to categorize the specific test(s) like so:
//
// var _ = Describe(testName, label.Blocking, func() { ... })
//
// Complex conditions can be provided to combine or negate different labels
// such as `ROSA && Upgrade` or `e2e && !privatelink`. See the [Ginkgo
// docs](https://onsi.github.io/ginkgo/#spec-labels) for more details. New
// labels introduced should be generic and applicable to multiple test suites.
//
// Labels can also be used in conjunction with the `--focus` flag to provide
// the ability to run a specific suite's category of tests, for example:
// `--label-filter Install --focus "Managed Cluster Validating Webhooks"`
// `--label-filter Upgrade --focus "Splunk Operator"`
package label

import "github.com/onsi/ginkgo/v2"

var (
	// Informing tests are new and needs to be proven stable before being
	// promoted to Blocking
	Informing = ginkgo.Label("Informing")

	// Blocking tests are stable and important enough to block a new release
	Blocking = ginkgo.Label("Blocking")

	// Install tests cover validating the component is available and ready
	Install = ginkgo.Label("Install")

	// Upgrade tests validate the component moving to a newer version
	Upgrade = ginkgo.Label("Upgrade")

	// ROSA tests support running on ROSA clusters
	ROSA = ginkgo.Label("ROSA")

	// HyperShift tests support running on a HyperShift cluster
	HyperShift = ginkgo.Label("HyperShift")

	// STS tests support running on a cluster deployed using STS
	STS = ginkgo.Label("STS")

	// PrivateLink tests support running on a cluster deployed using PrivateLink
	PrivateLink = ginkgo.Label("PrivateLink")

	// CCS tests support running on a Customer Cloud Subscription cluster
	CCS = ginkgo.Label("CCS")

	// E2E tests are included in a full end to end run of the suite
	E2E = ginkgo.Label("E2E")

	// AWS tests support running on a cluster in AWS
	AWS = ginkgo.Label("AWS")

	// GCP tests support running on a cluster in GCP
	GCP = ginkgo.Label("GCP")

	// Azure tests support running on a cluster in Azure
	Azure = ginkgo.Label("Azure")

	// Operator tests supported on all cluster types
	Operators = ginkgo.Label("Operators")

	// Service definition tests verifying openshift dedicated policies
	// https://docs.openshift.com/dedicated/osd_architecture/osd_policy/osd-service-definition.html
	ServiceDefinition = ginkgo.Label("ServiceDefinition")
)
