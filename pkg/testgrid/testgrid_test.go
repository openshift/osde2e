package testgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	testgrid "k8s.io/test-infra/testgrid/metadata"

	"github.com/openshift/osde2e/pkg/config"
)

const (
	junitReport = `
<testsuite failures="0" tests="2" time="2378.33176853">
<testcase classname="e2e.go" name="Extract" time="17.897214866"/>
<testcase classname="e2e.go" name="TearDown Previous" time="33.737494067"/>
</testsuite>`
	junitFileName = "junit_runner.xml"
)

func TestStartTestGridBuild(t *testing.T) {
	tg := setupTestGrid(t)

	// initiate a new build
	start := time.Now().UTC().Unix()
	ctx := context.Background()
	buildNum, err := tg.StartBuild(ctx, start)
	if err != nil {
		t.Fatalf("Could not start build: %v", err)
	}

	// check latest build has updated
	if curBuildNum, err := tg.getLatestBuild(ctx); err != nil {
		t.Errorf("Failed to get build number: %v", err)
	} else if curBuildNum != buildNum {
		t.Errorf("Current build (%d) does not match created build (%d)", curBuildNum, buildNum)
	}

	// confirm started file
	startedFile := new(testgrid.Started)
	if data, err := tg.getBuildFile(ctx, buildNum, startedFileName); err != nil {
		t.Errorf("Failed to get started file: %v", err)
	} else if err = json.Unmarshal(data, startedFile); err != nil {
		t.Errorf("Failed to decode started file: %v", err)
	} else if startedFile.Timestamp == 0 {
		t.Error("Timestamp was not set")
	}

	// write test results
	dir, err := ioutil.TempDir("", "osde2e-test")
	if err != nil {
		t.Fatalf("Failed to create result dir")
	}
	defer os.RemoveAll(dir)

	junitPath := filepath.Join(dir, junitFileName)
	if err = ioutil.WriteFile(junitPath, []byte(junitReport), os.ModePerm); err != nil {
		t.Fatalf("failed to write junit report: %v", err)
	}

	finish, passed := time.Now().UTC().Unix(), true
	finishedFile := testgrid.Finished{
		Timestamp: &finish,
		Passed:    &passed,
		Result:    "PASSED",
	}
	if err = tg.FinishBuild(ctx, buildNum, finishedFile, dir); err != nil {
		t.Fatalf("Failed to report results: %v", err)
	}

	// check for junit report
	name := filepath.Join(artifactsDir, junitFileName)
	if data, err := tg.getBuildFile(ctx, buildNum, name); err != nil {
		t.Errorf("Failed to get JUnit Report: %v", err)
	} else if !bytes.Equal([]byte(junitReport), data) {
		t.Error("Retrieved report does not match what was submitted")
	}

	// check finished file has been written
	actualFinished := new(testgrid.Finished)
	if data, err := tg.getBuildFile(ctx, buildNum, finishedFileName); err != nil {
		t.Errorf("Failed to get finished file: %v", err)
	} else if err = json.Unmarshal(data, actualFinished); err != nil {
		t.Errorf("Failed to decode finished file: %v", err)
	} else if actualFinished.Timestamp == nil {
		t.Error("Finished timestamp was nil")
	} else if *actualFinished.Timestamp != *finishedFile.Timestamp {
		t.Errorf("timestamp (%d) doesn't match expected (%d)", actualFinished.Timestamp, finishedFile.Timestamp)
	}
}

func setupTestGrid(t *testing.T) *TestGrid {
	cfg := config.Cfg
	checkTestGridEnv(t, cfg)

	tg, err := NewTestGrid(cfg.TestGridBucket, cfg.TestGridPrefix, []byte(cfg.TestGridServiceAccount))
	if err != nil {
		t.Fatalf("Failed setting up TestGrid: %v", err)
	}
	return tg
}

func checkTestGridEnv(t *testing.T, cfg *config.Config) {
	skipMsg := func(str string) {
		t.Skipf("The environment variable '%s' must be set to test TestGrid.", str)
	}

	if cfg.TestGridBucket == "" {
		skipMsg("TESTGRID_BUCKET")
	}

	if cfg.TestGridPrefix == "" {
		skipMsg("TESTGRID_PREFIX")
	}

	if len(cfg.TestGridServiceAccount) == 0 {
		skipMsg("TESTGRID_SERVICE_ACCOUNT")
	}

	if t.Skipped() {
		t.SkipNow()
	}
}
