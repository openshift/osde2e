# Region Enablement Job

- Ensure the region is enabled under the AWS account used for the job.
- Request SDA team (#service-delivery channel on slack) to enable the region for ocm account `rh-sd-cicd`. If there's an existing jira card for SDA to do this, mention it in the ping. 
- To run the job periodically, create a new job [in osde2e periodic jobs definition](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/osde2e/openshift-osde2e-main-periodics.yaml). Ensure that the desired region is specified as `CLOUD_PROVIDER_REGION` as one of the job's environment variables.  
  - OR you may trigger an ad hoc osde2e job per https://github.com/openshift/osde2e/blob/main/docs/Self-Service-MOPs/Ad-Hoc-E2E-Job.md