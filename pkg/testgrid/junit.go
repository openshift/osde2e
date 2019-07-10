package testgrid

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"k8s.io/test-infra/testgrid/metadata/junit"
)

const (
	// prefix all testsuites start with
	junitPrefix = "junit_"
)

// Suites returns the combined testsuites for the build.
func (t *TestGrid) Suites(ctx context.Context, buildNum int) (suites junit.Suites, err error) {
	suiteList, err := t.listSuites(ctx, buildNum)
	if err != nil {
		return suites, err
	}

	// combine into a single suites
	for _, suiteEntry := range suiteList {
		suites.Suites = append(suites.Suites, suiteEntry.Suites...)
	}
	return
}

func (t *TestGrid) listSuites(ctx context.Context, buildNum int) (suites []junit.Suites, err error) {
	// list all JUnit xml files for build
	prefix := fmt.Sprintf("%s/%s", ArtifactsDir, junitPrefix)
	junitPaths, err := t.ListFiles(ctx, buildNum, prefix, ".xml")
	if err != nil {
		return suites, fmt.Errorf("couldn't list JUnit Reports: %v", err)
	}

	for _, path := range junitPaths {
		// download suite data
		baseName := filepath.Base(path)
		data, err := t.getBuildFile(ctx, buildNum, ArtifactsDir, baseName)
		if err != nil {
			return suites, fmt.Errorf("failed getting suite data: %v", err)
		}

		// decode suites
		suite, err := junit.Parse(data)
		if err != nil {
			log.Printf("Failed to decode suite in '%s': %v", path, err)
			continue
		}

		// add to suites
		suites = append(suites, suite)
	}
	return suites, nil
}
