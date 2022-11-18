# Writing Tests

OSD end-to-end testing uses the [Ginkgo] testing framework and [Gomega]  matching libraries.

## Designing e2e Tests
We recommend the following basic principles in designing your e2e tests:
 - Map existing functionality to e2e test cases
 - Map any bugs found into e2e test cases 
 - Update e2e test cases as addon functionality is changed or added 
 - Ensure the assertions and logs output by tests reflect 
   - the execution paths and 
   - potential causes of failures  
   
Kubernetes guide on the topic https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/writing-good-e2e-tests.md 


## Writing first test

### Informing vs. Blocking
There are different suites of tests within OSDe2e with varying meanings and purposes. However, all new tests added should fall under the "informing" test suite.

This test suite is used as a proving ground for new tests to validate their quality and to ensure a potentially new flaky test does not impact the overall CI Signal.

Once a test has run for over a week with quality results, it can then be graduated into its correct/respective suite.

### Adding a new package of tests
All Ginkgo tests that are imported in **[`/cmd/osde2e/test/cmd.go`]** are ran as part of the osde2e suite.

For example, to add the tests in the Go package `github.com/openshift/osde2e/test/verify` to the normal test suite, you would add the following to **[`/cmd/osde2e/test/cmd.go`]**:
```go
import (
	_ "github.com/openshift/osde2e/pkg/e2e/verify"
)
```

### Adding a test to an existing package
This test from **[`/pkg/e2e/verify/imagestreams.go`]** provides a good example of setting up new ones:

- Create new file in a package that is imported by  **[`/cmd/osde2e/test/cmd.go`]** as discussed [above]. For this example, we will call the file **imagestreams.go**.

- Import [Ginkgo] testing framework and [Gomega] matching libraries:

**imagestreams.go**
```go
import (
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega
)
```

- Create new Describe block. These are used to organize tests into groups:

**imagestreams.go**
```go
var _ = ginkgo.Describe("[Suite: informing] ImageStreams", func() {
	// tests go here
})
```
**Note:** New tests must be initially added to the ["informing" test suite]. This allows existing signal to not be impacted by potentially flaky or unproven tests.

- Import the [helper package] and create new helper instance in Describe block. This will setup a Project for each test run and can be used to access the cluster.

**imagestreams.go**
```go
import (
	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: informing] ImageStreams", func() {
	h := helper.New()

	// tests go here
})
```

- Perform a request on the cluster. The helper provides various clientsets:

**imagestreams.go**
```go
import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = ginkgo.Describe("[Suite: informing] ImageStreams", func() {
	h := helper.New()

	list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
})
```

- Using [Gomega] the results of the request can be validated. The following checks that the requests to the cluster completed successfully and at least 50 ImageStreams exist cluster-wide:

**imagestreams.go**
```go
var _ = ginkgo.Describe("[Suite: informing] ImageStreams", func() {
	h := helper.New()

	util.GinkgoIt("should exist in the cluster", func() {
		list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list ImageStreams")
		Expect(list).NotTo(BeNil())

		numImages := len(list.Items)
		minImages := 50
		Expect(numImages).Should(BeNumerically(">", minImages), "need more images")
	})
})
```

All together this is a working and complete addition to the osde2e suite.

The "ImageStreams should exist in the cluster" test will run as part of the suite:

**imagestreams.go**
```go
package verify

import (
	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/common/helper"
)

var _ = ginkgo.Describe("[Suite: informing] ImageStreams", func() {
	h := helper.New()

	util.GinkgoIt("should exist in the cluster", func() {
		list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred(), "couldn't list ImageStreams")
		Expect(list).NotTo(BeNil())

		numImages := len(list.Items)
		minImages := 50
		Expect(numImages).Should(BeNumerically(">", minImages), "need more images")
	})
})
```

## Ginkgo

### Setup & Teardown
[Ginkgo] has been configured to bring a cluster up in it's [`BeforeSuite`] and destroy it in it's [`AfterSuite`].

**Cluster configuration**
- Launched clusters are setup with the [osd package]
	- Changes to the way test clusters are launched should be made there
- [ocm-sdk-go] is used to launch clusters
- Configuration for launching clusters is loaded from a [`config.Config`] instance

## Helper
A helper can be created in tests using [`helper.New()`]

The helper:
- Configures Ginkgo to create a Project before each test and delete it after
- Provides access to OpenShift and Kubernetes clients configured for the test cluster
- Provides commonly used test functions

## Static files

Static files for `OSDe2e`  such as YAML manifests are managed using the native
Go embed feature. If your test has a static asset such as a manifest, add it
into the **[`/assets/`]** directory. It can then be accessed through the
`assets.FS` [embedded filesystem](https://pkg.go.dev/embed#hdr-File_Systems)

[Ginkgo]:https://onsi.github.io/ginkgo/
[Gomega]:https://onsi.github.io/gomega/
[`/cmd/osde2e/test/cmd.go`]:/cmd/osde2e/test/cmd.go
[above]:#adding-a-new-package-of-tests
[`/pkg/e2e/verify/imagestreams.go`]:/pkg/e2e/verify/imagestreams.go
["informing" test suite]:/configs/informing-suite.yaml
[helper package]:/pkg/common/helper/
[osd package]:/pkg/common/osd
[`BeforeSuite`]:https://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite
[`AfterSuite`]:https://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite
[ocm-sdk-go]:https://github.com/openshift-online/ocm-sdk-go
[`config.Config`]:https://godoc.org/github.com/openshift/osde2e/common/pkg/config#Config
[`helper.New()`]:https://godoc.org/github.com/openshift/osde2e/pkg/common/helper#New
[`/assets/`]:/assets/
[Hive]:https://github.com/openshift/hive
