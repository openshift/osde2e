# Test Harnesses

Test harnesses are standalone ginkgo e2e test images run on test pods on test clusters by osde2e framework. There are two types of test harnesses depending on the target component tested. SOPs to create and run both are as described below.
- [Operator Test Harness](#operator-test-harness)
- [Addon Test Harness](#addon-test-harness)
- [Generic Test Harness](#generic-test-harness)

## Operator Test Harness
1. Clone operator repo.

2. Add subscription to harness boilerplate.
   Add `openshift/golang-osd-operator-osde2e` line to /boilerplate/update.cfg

3. Run `make boilerplate-update`. Make a commit with updated bp. This commit must be separate from other code changes, otherwise boilerplate pr-check will fail.

4. Run `make e2e-harness-generate`.

5. Update package names in each test file appropriately for the new directory structure.

5. Write tests under osde2e/ generated test files.

6. If you're migrating an existing tests from osde2e repo: Copy test files from osde2e/pkg/e2e/operators/<your-operator> to  the newly generated /osde2e directory in your operator.

7. Run `make e2e-harness-build`
   If you have go mod errors, `make e2e-harness-build` target won't work as it includes go mod tidy. If your repo uses go versions prior to 1.18, you’d need to manually run the following to build a test file
   `go test ./osde2e -v -c --tags=integration -o harness.test`

Any dependency errors must be resolved to create the test binary.

9. Run `make e2e-image-build-push` to push harness docker image. Default operator quay repos are under app-sre quay org. If you wish to push to your dev repo, provide registry and repo to this command: `REGISTRY_TOKEN=xx HARNESS_IMAGE_REGISTRY=quay.io HARNESS_IMAGE_REPOSITORY=<dev-repo> make e2e-image-build-push`

10. Test this image locally with `osde2e` binary.
	For testing with osde2e, use [example harness SOP](https://github.com/ritmun/osde2e-example-test-harness#locally-running-your-test-harness)
	- For `TEST_HARNESSES` env var, provide url to the quay image pushed in #9 above.
	- Do not provide `ADDON_IDS`.
    
11. Run unit tests on operator repo to ensure updated code doesn't break existing functionality: `make test`

12. Create PR against operator repo, watch pr-checks.

13. Automate test harness publishing:
	Add  `make e2e-image-build-push`  target in the build pipeline for your operator in `app-interface` repo  similar to this in AVO:
	https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/osd-operators/cicd/ci-int/jobs-aws-vpce-operator.yaml#L27

14. Automate e2e test harness testing: To run `latest` harness image as a postsubmit job: Add a test step to the following to the prow config file in release repo using the example of [`stage-e2e-harness` job config in AVO](  https://github.com/openshift/release/blob/b6f9d2c0bffaa230a8097fb97d5abb4e91f96e4d/ci-operator/config/openshift/aws-vpce-operator/openshift-aws-vpce-operator-main.yaml). 
    - Run `make ci-operator-config`
    - Run `make jobs`
    - Commit updated job and create a PR against `openshift/release`  repo

15. To run test harness as a periodic job : (todo  https://issues.redhat.com/browse/SDCICD-955)

## Addon Test Harness

1. This method is recommended when the target component is an openshift addon. 
2. Follow the example harness SOP [here](https://github.com/openshift/osde2e-example-test-harness)
   - Note the addon id environment parameter "ADDON_IDS" mentioned in the SOP. This should be provided to osde2e executable for it to install your addon prior to test run.

## Generic Test Harness

1. This method is recommended to be used for any component which does not use openshift operator boilerplate convention. It requires manual creation of test structure in either the target repository or a standalone one.

2. Follow the [example test harness](https://github.com/openshift/osde2e-example-test-harness) structure to create a test harness in your component's repository or a standalone repository.

3. Follow steps 7 onwards under [Operator test harness](#operator-test-harness) SOP above to publish and run your test harness.
