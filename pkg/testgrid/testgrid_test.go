package testgrid

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	testgrid "k8s.io/test-infra/testgrid/metadata"

	"github.com/openshift/osde2e/pkg/config"
)

func TestStartTestGridBuild(t *testing.T) {
	tg := setupTestGrid(t)

	// initiate a new build
	now := time.Now().UTC()
	ctx := context.Background()
	buildNum := startBuild(t, ctx, tg, now)

	// confirm started file
	if startedFile, curBuildNum, err := tg.LatestStarted(ctx); err != nil {
		t.Errorf("Failed to get started record: %v", err)
	} else if curBuildNum != buildNum {
		t.Errorf("Current build (%d) does not match created build (%d)", curBuildNum, buildNum)
	} else if startedFile.Timestamp == 0 {
		t.Error("Timestamp was not set")
	}

	// write test results
	dir, junitFileName, suiteData := writeTestSuite(t, "")
	defer os.RemoveAll(dir)

	finish, passed := time.Now().UTC().Unix(), true
	finishedFile := testgrid.Finished{
		Timestamp: &finish,
		Passed:    &passed,
		Result:    "PASSED",
	}
	if err := tg.FinishBuild(ctx, buildNum, &finishedFile, dir); err != nil {
		t.Fatalf("Failed to report results: %v", err)
	}

	// check for junit report
	name := filepath.Join(ArtifactsDir, junitFileName)
	if data, err := tg.getBuildFile(ctx, buildNum, name); err != nil {
		t.Errorf("Failed to get JUnit Report: %v", err)
	} else if !bytes.Equal(suiteData, data) {
		t.Error("Retrieved report does not match what was submitted")
	}

	// check finished file has been written
	if actualFinished, err := tg.Finished(ctx, buildNum); err != nil {
		t.Errorf("Failed to get finished file: %v", err)
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
		t.Fatalf("Failed setting up TestGrid: %v ", err)
	}
	return tg
}

func startBuild(t *testing.T, ctx context.Context, tg *TestGrid, when time.Time) (buildNum int) {
	started := testgrid.Started{
		Timestamp: when.Unix(),
	}
	buildNum, err := tg.StartBuild(ctx, &started)
	if err != nil {
		t.Fatalf("Could not start build: %v", err)
	}
	return buildNum
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
