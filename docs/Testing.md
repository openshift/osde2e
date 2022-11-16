# **Testing**    

New versions of OpenShift must be qualified as part of a continuous delivery approach into managed environments.  The OpenShift Dedicated End to End (osde2e) test framework facilitates this for two primary use-cases:
* Managed OpenShift (OSD, ROSA, ARO).  osde2e test results are part of the gating signal for promotion between environments.
* OSD Operators that run on top of Managed OpenShift. See https://github.com/openshift/osde2e#operator-testing
* **Addons** that run on top of Managed OpenShift.  Integration testing of two pieces of software (the Addon and the version of OCP it will run on) gives Addon owners the earliest possible signal as to whether newer versions of OpenShift (as deployed in OSD, ROSA or ARO) will affect their software.  This gives Addons owners time to fix issues well in advance of release. Please refer to the [addon test harness docs](https://github.com/openshift/osde2e-example-test-harness/blob/main/README.md) for SOP.

