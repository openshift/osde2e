# Region Enablement Job

- Ensure the region is enabled under the AWS account used for the job (in prow: 159042463696, in jenkins:652144585153).
- Request SDA team (#service-delivery channel on slack) to enable the region for ocm account `rh-sd-cicd`. If there's an existing jira card for SDA to do this, mention it in the ping. 
- To run the job, you may create a release repo PR Create a new PR for the release repo osde2e jobs with a periodic job specifying the desired region
- OR you may run an ad hoc osde2e job per https://github.com/openshift/osde2e/blob/main/docs/Self-Service-MOPs/Ad-Hoc-E2E-Job.md