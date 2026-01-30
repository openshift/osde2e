# Testing with OSDE2E

New versions of OpenShift must be qualified as part of a continuous delivery
approach into managed environments. The OpenShift Dedicated End to End (osde2e)
test framework facilitates this for the following primary use-cases:

* Managed OpenShift (OSD, ROSA, ROSA HCP, ARO)
  * OSDe2e test results are part of the gating signal for promotion between environments
* OSD Operators that run on top of Managed OpenShift
* **Addons** that run on top of Managed OpenShift. Integration testing of two pieces
  of software (the Addon and the version of OCP it will run on) gives Addon owners
  the earliest possible signal as to whether newer versions of OpenShift
  (as deployed in OSD, ROSA or ARO) will affect their software. This gives Addons
  owners time to fix issues well in advance of release. Please refer to the
  [addon test harness docs][Test Harness Repo] for SOP (Standard Operational Procedure).

[Test Harness Repo]: https://github.com/openshift/osde2e-example-test-harness/blob/main/README.md

- [Self Service](#self-service)
	- [Instance Type Enablement Testing](#instance-type-enablement)
	- [Region Enablement Testing](#region-enablement)
	- [Ad-Hoc E2E test run](#ad-hoc-e2e-test-run)
- [E2E Test for OSD Operators](#e2e-test-for-osd-operators)
- [Executing Monorepo Osde2e Tests](#executing-monorepo-osde2e-tests)

# Instance Type Enablement
- AWS account used for the job (in prow: 159042463696, in jenkins:652144585153).
- SOP: https://github.com/openshift/osde2e/blob/main/docs/Self-Service-MOPs/Instance-Type-Enablement.md

# Region Enablement
- AWS account used for the job (in prow: 159042463696, in jenkins:652144585153).
- Follow these steps https://github.com/openshift/osde2e/blob/main/docs/Self-Service-MOPs/Region-Enablement.md

# Ad-Hoc e2e test run
- Uses AWS account 652144585153
- [Follow the SOP here](https://github.com/openshift/osde2e/blob/main/docs/adhoc-osde2e-testing.md)

# E2E Test for OSD Operators
- [E2E Test SOP](https://github.com/openshift/ops-sop/blob/master/v4/howto/osde2e/operator-test-harnesses.md)

# Executing Monorepo Osde2e Tests
- [Running Osde2e SOP](https://github.com/openshift/osde2e/blob/main/docs/run-osde2e-tests.md)
