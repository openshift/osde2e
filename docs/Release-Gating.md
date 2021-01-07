# Gate Releases With CI/CD Data

OSDe2e enables users to procure CICD data with the help of prometheus queries. These queries return the summary of a specific job that has run on a cluster. The final pass value check can be used to determine if gating can be dealt with.

## General Example

A simple query command would be as follows:

```
./osde2e query "cicd_jUnitResult{cloud_provider=\"aws\", environment=\"prod\", cluster_id=\"example-id\"}" --output-format json
```


The above query would return results that are jUnitResult metrics associated with 'aws' as the cloud provider, production as the cluster environment and the corresponding cluster with the cluster-id 'example-id'. Multiple results may be displayed and these can be filtered out with additional parameters while querying for results.


Results would be published in the following manner:


    "metric": {
      "__name__": "cicd_jUnitResult",
      "cloud_provider": "aws",
      "cluster_id": "example-id",
      "endpoint": "scrape",
      "environment": "prod",
      "install_version": "openshift-v4.5.16",
      "job": "osde2e-prod-rosa-e2e-default",
      "job_id": "1330662521106862080",
      "namespace": "app-sre-observability-production",
      "phase": "install",
      "pod": "pushgateway-3-fc4vb",
      "region": "eu-central-1",
      "result": "skipped",
      "service": "pushgateway-nginx-gate",
      "suite": "OSD e2e suite",
      "testname": "[install] [Suite: scale-performance] Scaling should be tested with HTTP"
    },
    "value": [
      1606097115.972,
      "0.620441756"
    ]
  }


Each of these results have a value parameter which contains a pass check ratio which reflects passed cases against failed ones. 

Similarly, other queries can be given to procure different results.

## Other sample queries

These values can be plugged in as query search strings while running the `./osde2e query ..... --output-format json` command.

### Count queries

Query search string

Usage
`count by (job) (cicd_jUnitResult)`
This will give a list of all of the osde2e jobs in the form of jUnitResult metrics
`count by (job_id) (cicd_jUnitResult{job=\"dummyjob\"})`
This will list all of the individual job IDs (individual job runs) for a given job
`count by (cluster_id) (cicd_jUnitResult{cloud_provider=\"aws\", environment=\"prod\"})`
This will list all of the individual cluster IDs for an provider and environment


In the above pattern, we can find the count for any parameter similar to 'job' or job_id' as seen above in the parameters returned by the sample result along with filters.

To get a pass ratio of results that passed or were skipped against the total number of results, users can make use of query search strings as seen below - 

`
count by (install_version) (cicd_jUnitResult{cloud_provider=\"aws\",install_version=\"openshift-v4.6.4\", environment=\"prod\", result=~\"passed|skipped\"}) / count by (install_version) (cicd_jUnitResult{cloud_provider=\"aws\",install_version=\"openshift-v4.6.4\", environment=\"prod\"})
`

The above search string would return the pass ratio of the passed/skipped results against total results for clusters installed with openshift version 4.6.4 in prod under aws.

### Event Metric queries

For procuring event metric results, switching cicd_jUnitResult with cicd_event in the query string along with the usual filters would be enough as seen below.

`cicd_event{cloud_provider=\"aws\", environment=\"prod\", cluster_id=\"example-id\"}`


### Metadata Metric queries

For procuring metadata metric results, using cicd_metadata in the query string along with the usual filters would be enough as seen below.

`cicd_metadata{cloud_provider=\"aws\", environment=\"prod\"}`


### Addon Metadata Metric queries

For procuring addon metadata metric results, using cicd_metadata in the query string along with the usual filters would be enough as seen below.

`cicd_addon_metadata{job=\"sample-job\", job_id=\"sample-job-id\"}`


### Queries to search for results containing parameter value

Filters can be used to return results with parameters that contain a specific value. The query string below is an example.

`cicd_jUnitResult{result=\"failed\", testname=~\".*prow-job.*\"}`

The above query would return results that have failed along with testnames that contain the string 'prow-job' in-between the actual string.

`cicd_jUnitResult{result=~\"passed|skipped\"}`

The above query would return results that either contain the string 'passed' or 'skipped'.

### Version check queries

The osde2e query command also allows users to get the pass check values for installed versions and upgrade versions for a cluster. To do this, enabling the -version check bool option in the osde2e command would be enough along with the version parameters as seen below.

```
./osde2e query -v -i openshift-v4.6.0 -u openshift-v4.6.4
```

This would return the pass ratio of results that passed or were skipped against the total number of results. 


## Using the metrics package

The metrics package `github.com/openshift/osde2e/pkg/metrics` in osde2e can also be used to query results by utilizing some of its helper functions which take in query search strings and a time range as input.
More details can be found in the file: https://github.com/openshift/osde2e/blob/main/docs/Metrics-Client.md


## Using scripts

Users can make use of custom scripts to extract the pass percentage check for a given job/pipeline and see if it is below a given threshold to decide if the data must be gated or not.

An example would be as follows:
```
#!/bin/bash
#
# Check if a query result are lower than a defined threshold
#
 
while getopts t:c: option
do
case "${option}"
in
t) THRESHOLD=${OPTARG};;
c) CLUSTER_ID=${OPTARG};;
esac
done
 
cmd=$(./osde2e query "cicd_event{cloud_provider=\"aws\", environment=\"prod\", cluster_id=\"$CLUSTER_ID\"}" --output-format json 2>&1 | tail -4 | head -1 | xargs)
 
 
echo $cmd
 
 
if (( $(echo "$cmd <= $THRESHOLD" | bc -l) )); then
   exit 1
fi

The above script runs a jUnitresult query against a cluster with ‘sample-cluster-id’ as the cluster-id along with the other filters. This is to check a single result and would need more parameters to filter it out as this script would extract the value from the end of the file.
```
The script can be run as:

```
$ ./demo.sh -t 2 -c sample-cluster-id
1
```

Here, the output displays the pass ratio which seems to be 1 in this case. It would just exit if the value is lower than the threshold.

The `-t` option is to set a threshold value and the cmd variable would hold the query command. Depending on the query, users can modify the variable to extract the value. Similarly, other options can be given for different query strings.

For the jUnitResult, `tail -4` and `head -1` are used to extract the value at the 4th line from the last of the returned result. Note that the value might be at a different line for a different query string and this means that the tail value has to be changed accordingly.

For version checks, users can use the command:

```
./osde2e query -v -i openshift-v4.6.4 2>&1 | grep version | awk -F: '{split($0, a,"-"); print a[2]}'
```

There’s a sample script below.

```
$ ./demo.sh -t 2 -i openshift-v4.6.4
0.9900793650793651
```
```
#!/bin/bash
#
# Check if a query result are lower than a defined threshold
#
 
while getopts t:i: option
do
case "${option}"
in
t) THRESHOLD=${OPTARG};;
i) INSTALL_VERSION=${OPTARG};;
 
esac
done
 
 
cmd=$(./osde2e query -v -i $INSTALL_VERSION 2>&1 | grep version | awk -F: '{split($0, a,"-"); print a[2]}')
 
 
echo $cmd
 
 
 
if (( $(echo "$cmd <= $THRESHOLD" | bc -l) )); then
   exit 1
fi
```

Here, you can provide the install version as an option and plug it in the query string as seen above. You can also give an upgrade version option if that is needed in the query search string.
