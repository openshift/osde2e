package prow

import (
	"fmt"
	"os"
)

// JobURL infers the URL of this job using environment variables
// provided by Prow. It is not foolproof, and the URLs generated
// are only valid for "JOB_TYPE=periodic" jobs.
func JobURL() (url string, ok bool) {
	if os.Getenv("JOB_TYPE") != "periodic" {
		return
	}
	var jobID, jobName string
	if jobID, ok = os.LookupEnv("BUILD_ID"); !ok {
		return
	}
	if jobName, ok = os.LookupEnv("JOB_NAME"); !ok {
		return
	}
	return fmt.Sprintf("https://prow.ci.openshift.org/view/gs/origin-ci-test/logs/%s/%s", jobName, jobID), true
}
