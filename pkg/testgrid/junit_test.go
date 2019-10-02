package testgrid

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	testgrid "k8s.io/test-infra/testgrid/metadata"
)

const (
	junitSuiteText = `
<testsuite failures="0" tests="2" time="2378.33176853">
<testcase classname="e2e.go" name="Extract" time="17.897214866"/>
<testcase classname="e2e.go" name="TearDown Previous" time="33.737494067"/>
</testsuite>`
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestListSuites(t *testing.T) {
	tg := setupTestGrid(t)

	// initiate a new build
	now := time.Now().UTC()
	ctx := context.Background()
	buildNum := startBuild(t, ctx, tg, now)

	// setup writing testsuites
	count := 5
	dir := ""
	for i := 0; i < count; i++ {
		dir, _, _ = writeTestSuite(t, dir)
	}
	defer os.RemoveAll(dir)

	// upload results
	finishTimestamp := time.Now().UTC().Unix()
	passed := true
	finishedFile := testgrid.Finished{
		Timestamp: &finishTimestamp,
		Passed:    &passed,
		Result:    "PASSED",
	}
	if err := tg.FinishBuild(ctx, buildNum, &finishedFile, dir); err != nil {
		t.Fatalf("Failed to report results: %v", err)
	}

	// get suites
	suites, err := tg.Suites(ctx, buildNum)
	if err != nil {
		t.Fatalf("Error retrieving suites for build %d: %v", buildNum, err)
	}

	// count test occurrence
	testCount := map[string]int{}
	for _, s := range suites.Suites {
		for _, r := range s.Results {
			curCount := testCount[r.Name]
			testCount[r.Name] = curCount + 1
		}
	}

	// check if counts match
	for testName, c := range testCount {
		if c != count {
			t.Fatalf("test count for '%s' doesn't match: have %d, wanted %d", testName, c, count)
		}
	}
}

func writeTestSuite(t *testing.T, dirIn string) (dir, name string, suiteData []byte) {
	var err error

	// if dirIn empty, create temporary dir
	if dirIn == "" {
		dir, err = ioutil.TempDir("", "osde2e-test")
		if err != nil {
			t.Fatalf("Failed to create result dir: %v", err)
		}
	} else {
		dir = dirIn
	}

	// set report
	name = fmt.Sprintf("junit_%s.xml", randomStr(4))
	suiteData = []byte(junitSuiteText)

	// write file
	junitPath := filepath.Join(dir, name)
	if err = ioutil.WriteFile(junitPath, suiteData, os.ModePerm); err != nil {
		t.Fatalf("failed to write junit report: %v", err)
	}
	return
}

func randomStr(length int) (str string) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < length; i++ {
		c := string(chars[rand.Intn(len(chars))])
		str += c
	}
	return
}
