# Writing Tests

OSD end-to-end testing uses the [Ginkgo](https://onsi.github.io/ginkgo/) testing framework and [Gomega](https://onsi.github.io/gomega/) matching libraries.

## Writing first test

### Adding a new package of tests
All Ginkgo tests that are imported in **[e2e_test.go](../e2e_test.go)** are ran as part of the osde2e suite.

For example, to add tests from the Go package `github.com/openshift/osde2e/test/verify` to the suite add the following to **[e2e_test.go](../e2e_test.go)**:
```go
import (
	_ "github.com/openshift/osde2e/test/verify"
)
```

### Adding a test to an existing package
This test from **[./test/verify/imagestreams.go](../test/verify/imagestreams.go)** provides a good example of setting up new ones:

- Create new file in a package that is imported by **[e2e_test.go](../e2e_test.go)** as discussed [above](#adding-a-new-package-of-tests). For this example, we will call the file **imagestreams.go**.

- Import [Ginkgo](https://onsi.github.io/ginkgo/) testing framework and [Gomega](https://onsi.github.io/gomega/) matching libraries:

**imagestreams.go**
```go
import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega
)
```

- Create new Describe block. These are used to organize tests into groups:

**imagestreams.go**
```go
var _ = ginkgo.Describe("ImageStreams", func() {
	// tests go here
})
```

- Import helper package and create new helper instance in Describe block. This will setup a Project for each test run and can be used to access the cluster.

**imagestreams.go**
```go
import (
	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("ImageStreams", func() {
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

var _ = ginkgo.Describe("ImageStreams", func() {
	h := helper.New()

	list, err := h.Image().ImageV1().ImageStreams(metav1.NamespaceAll).List(metav1.ListOptions{})
})
```

- Using Gomega the results of the request can be validated. The following checks that the requests to the cluster completed successfully and at least 50 ImageStreams exist cluster-wide:

**imagestreams.go**
```go
var _ = ginkgo.Describe("ImageStreams", func() {
	h := helper.New()

	ginkgo.It("should exist in the cluster", func() {
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

The "ImageStreams should exist in the cluster" test will run as part of the suite and have its results uploaded to TestGrid:

**imagestreams.go**
```go
package verify

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/osde2e/pkg/helper"
)

var _ = ginkgo.Describe("ImageStreams", func() {
	h := helper.New()

	ginkgo.It("should exist in the cluster", func() {
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
Ginkgo has been configured to bring a cluster up in it's [`BeforeSuite`](https://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite) and destroy it in it's [`AfterSuite`](https://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite).

**Cluster configuration**
- Launched clusters are setup with the [`osd`](../pkg/osd) package
	- Changes to the way test clusters are launched should be made there
- [ocm-sdk-go](https://github.com/openshift-online/ocm-sdk-go) is used to launch clusters
- Configuration for launching clusters is loaded from a [`config.Config`](https://godoc.org/github.com/openshift/osde2e/pkg/config#Config) instance

## Helper
A helper can be created in tests using [`helper.New()`](https://godoc.org/github.com/openshift/osde2e/pkg/helper#New).

The helper:
- Configures Ginkgo to create a Project before each test and delete it after
- Provides access to OpenShift and Kubernetes clients configured for the test cluster
- Provides commonly used test functions

## TestGrid
Results of tests are uploaded to an instance of [TestGrid](https://testgrid.k8s.io/redhat-openshift-release-blocking) to allow analysis. All logs provided through the OSD API are additionally uploaded.

TestGrid is configured through [`config.Config`](https://godoc.org/github.com/openshift/osde2e/pkg/config#Config).
