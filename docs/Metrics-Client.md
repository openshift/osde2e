# Metrics Client

OSDe2e produces and sends many metrics up to Prometheus managed by the App-SRE team. To enable teams to consume this data in a simpler fashion, OSDe2e has a metrics consumption library bundled with commonly-used queries built in.


## Prometheus Queries

Access to Prometheus has to be granted by the App-SRE team and this can be done by sending a pull request as indicated by the instructions shown here - [app-sre monitoring](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/docs/app-sre/monitoring.md#prometheus)


Once access is granted, users can directly query in app-sre prometheus and view results through the following link - [app-sre prometheus](https://prometheus.app-sre-prod-01.devshift.net/).




## Quickstart Example

```golang
import (
    "github.com/openshift/osde2e/pkg/metrics"
    "log"
    "time"
)

func main(){
    // NewClient returns a new metrics client.
    // If no arguments are supplied, the global config will be used.
    // You can set PROMETHEUS_ADDRESS and PROMETHEUS_BEARER_TOKEN environment variables for the global config.
    // If one argument is supplied, it will be used as the address for Prometheus, but will use the global config for the bearer token.
    // If two arguments are supplied, the first will be used as the address for Prometheus and the second will be used as the bearer token.
    client, err := metrics.NewClient()
    if err != nil {
        log.Errorf("Error creating metrics client: %s", err.Error())
    }

    // Specify a job name to look up
    job := "osde2e-prod-rosa-e2e-default"

    end := time.Now()
    // Look back 24 hours
	start := end.Add(-time.Hour * 24)

    results, err := client.ListJUnitResultsByJobName(job, start, end)
    if err != nil {
        log.Errorf("Error getting JUnit results: %s", err.Error())
    }

    log.Println("JobID/Phase, JobName, Result, Duration")
    for _, result := results {
        log.Printf("%s/%s, %s, %s, %d", result.JobID, result.Phase, result.TestName, result.Result, result.Duration)
    }
}

```

## API Docs

There are many different helper functions to abstract away common queries. To view these methods and the metrics data structures, see [https://godoc.org/github.com/openshift/osde2e/pkg/metrics](https://godoc.org/github.com/openshift/osde2e/pkg/metrics)

