package testgrid

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
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
	suitePrefix := t.buildFileKey(buildNum, artifactsDir, junitPrefix)
	suiteIt := t.bucket.Objects(ctx, &storage.Query{
		Prefix:    suitePrefix,
		Delimiter: "/",
	})

	for {
		obj, err := suiteIt.Next()
		// stop when done, return errs, and skip non-XML
		if err == iterator.Done {
			break
		} else if err != nil {
			return suites, err
		} else if !strings.HasSuffix(obj.Name, ".xml") {
			continue
		}

		// download suite data
		baseName := filepath.Base(obj.Name)
		data, err := t.getBuildFile(ctx, buildNum, artifactsDir, baseName)
		if err != nil {
			return suites, fmt.Errorf("failed getting suite data: %v", err)
		}

		// decode suites
		suite, err := junit.Parse(data)
		if err != nil {
			log.Printf("Failed to decode suite in '%s': %v", obj.Name, err)
			continue
		}

		// add to suites
		suites = append(suites, suite)
	}
	return suites, nil
}
