# Region Enablement Job

- Ensure the region is enabled under the AWS account used for the job.
- Request SDA team (#service-delivery channel on slack) to enable the region for ocm account `rh-sd-cicd`. If there's an existing jira card for SDA to do this, mention it in the ping. 
- To run the job periodically on prow, create a new job [in osde2e periodic jobs definition](https://github.com/openshift/release/blob/master/ci-operator/jobs/openshift/osde2e/openshift-osde2e-main-periodics.yaml). Ensure that the desired region is specified as `CLOUD_PROVIDER_REGION` as one of the job's environment variables.  
- To run the tests as an ad-hoc jenkins job, [follow the SOP here](https://github.com/openshift/ops-sop/blob/master/v4/howto/osde2e/adhoc-jenkins-testing.md) 

### Frequent Issues

-  For reference, this doc describes all the steps that need to be done before osde2e should be triggered. Frequently, issues incurred by osde2e are due to these steps being missed. https://docs.google.com/document/d/1jFDa3LB1vmWCnS_G8G_1xGb8Ch4iJgHc_BrR6OsRVbs/edit#heading=h.f5kol9rb0w93 . More frequent issues are described below:
-  You may get an AMI related error similar to: `ERR: Failed to create cluster: There is no AMI available for version openshift-v4.12.16 on region me-central-1`
  - If you get this, report to BU. They should create a jira card similar to this [OHSS-22229](https://issues.redhat.com/browse/OHSS-22229) 
- You may get quota errors similar to this: `ERR: Failed to create cluster: required total number of vCPU quota for install is '40': '24' vCPU for control plane nodes, '8' vCPU for infra nodes and '8' vCPU for compute nodes, which exceeds the available vCPU quota of '5'` 
  - If you get such an error, log in to your aws account's console, navigate to the [quota request page](https://me-central-1.console.aws.amazon.com/servicequotas/home/services/ec2/quotas/L-34B43A08), and request quota increase. The default is typically '5'. We have requested a quota of '384' in the past for region enablement jobs. 



