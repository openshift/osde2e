# Writing Tests

This page provides you with resources/references to write tests
to validate various aspects of Managed OpenShift clusters.

Tests can be consumed by osde2e test framework in one of two ways:

1. Runs tests outside of osde2e using its test harness runner (recommended)
2. Runs tests that reside within osde2e (legacy mode/not recommended)

For any new tests, they should be written in which osde2e consumes them using
its test harness runner feature. *This applies to all OSD SREP operator tests.*

To learn more about the test harness runner of osde2e, refer to the following
[document](Test-Harnesses.md).

## Test Standards "Best Practices"

When writing new tests or enhancing existing tests, every test should strive
to comply to the following standards:

* Follow the [Kubernetes best practices guide] when writing end to end tests.
* Review [Ginkgo] and [Gomega] documentation as these are the core test
  frameworks when writing OSD SREP operator tests.
* Use [osde2e-common] module as much as possible when writing test cases. This
  module provides common modules when working with Managed OpenShift which
  aim to reduce code duplication across tests such as
  clients for interfacing with OCM, OpenShift, Prometheus and more.
* Use the [e2e-framework] as much as possible and become familiar with it
  when interfacing with OpenShift clusters.
* Apply labels "tags" to your test cases allowing for easy classification
  "grouping" of tests. This is helpful when certain tests would like to be run
  over the entire test suite.
* Keep test cases focused on their specific scope. Test cases are best to be
  mapped to a given feature/functionality for the product or OSD SREP operator.
* Ensure both positive/negative cases are covered for your test case.
* Ensure every test case has proper error messages logged with using [Gomega]
  matchers. This helps when it comes to
  [troubleshooting/debugging failing tests][debugging tests].

## Examples

You can find well defined examples of existing tests following the test
standards mentioned above below. The examples are showing how tests are written
for OSD SREP operator tests:

* [Managed Upgrade Operator Tests][managed-upgrade-operator-tests]
* [OCM Agent Operator Tests][ocm-agent-operator-tests]
* [RBAC Permissions Operator Tests][rbac-operator-tests]

Additional examples can be found below:

* [Cluster Difference Test Suite][cluster-diff-test-suite]
* [Management/Service Cluster Upgrade Test Suite][mc-sc-upgrade-testsuite]

[cluster-diff-test-suite]: https://github.com/openshift/osde2e/blob/main/test/cluster_diff/cluster_diff_test.go
[debugging tests]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/writing-good-e2e-tests.md#debuggability
[e2e-framework]: https://github.com/kubernetes-sigs/e2e-framework
[Ginkgo]:https://onsi.github.io/ginkgo/
[Gomega]:https://onsi.github.io/gomega/
[Kubernetes best practices guide]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/writing-good-e2e-tests.md
[managed-upgrade-operator-tests]: https://github.com/openshift/managed-upgrade-operator/blob/master/osde2e/managed_upgrade_operator_tests.go
[mc-sc-upgrade-testsuite]: https://github.com/openshift/osde2e/blob/main/test/mcscupgrade/mcscupgrade_test.go
[ocm-agent-operator-tests]: https://github.com/openshift/ocm-agent-operator/blob/master/osde2e/ocm_agent_operator_tests.go
[osde2e-common]: https://github.com/openshift/osde2e-common
[rbac-operator-tests]: https://github.com/openshift/rbac-permissions-operator/blob/master/osde2e/rbac_permissions_operator_tests.go
