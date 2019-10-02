// Package testgrid provides a client for reporting build results to a TestGrid instance.
package testgrid

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	testgrid "k8s.io/test-infra/testgrid/metadata"
)

// NewTestGrid for bucket and prefix using b64ServiceAccount (base64 encoded GCloud Service Account JSON key) for auth.
// Google Cloud Service Account related to b64ServiceAccount must have read/write access to bucket.
//
// A b64ServiceAccount can be obtained through multiple methods documented here:
// https://cloud.google.com/iam/docs/creating-managing-service-account-keys#creating_service_account_keys
func NewTestGrid(bucket, prefix string, b64ServiceAccount []byte) (*TestGrid, error) {
	if bucket == "" {
		return nil, errors.New("bucket for TestGrid is not set")
	} else if prefix == "" {
		return nil, errors.New("prefix for TestGrid is not set")
	} else if len(b64ServiceAccount) == 0 {
		return nil, errors.New("a Service Account for TestGrid is not set")
	}

	serviceAccount := make([]byte, base64.StdEncoding.DecodedLen(len(b64ServiceAccount)))
	if _, err := base64.StdEncoding.Decode(serviceAccount, b64ServiceAccount); err != nil {
		return nil, fmt.Errorf("could not base64 decode Service Account JSON: %v", err)
	}

	ctx := context.Background()
	gcsClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(serviceAccount))
	if err != nil {
		return nil, err
	}

	return &TestGrid{
		bucket: gcsClient.Bucket(bucket),
		prefix: prefix,
	}, nil
}

// TestGrid allows reporting to a TestGrid instance.
type TestGrid struct {
	// instance options
	bucket *storage.BucketHandle
	prefix string
}

// StartBuild uploads started and updates the latest build to point to it.
func (t *TestGrid) StartBuild(ctx context.Context, started *testgrid.Started) (buildNum int, err error) {
	curBuildNum, err := t.getLatestBuild(ctx)
	if err != nil {
		return 0, fmt.Errorf("couldn't get latest build number: %v", err)
	}

	// increment for current build
	buildNum = curBuildNum + 1
	if err = t.setLatestBuild(ctx, buildNum); err != nil {
		return buildNum, fmt.Errorf("couldn't update latest build number: %v", err)
	}

	// upload started file
	if err = t.writeBuildFile(ctx, buildNum, startedFileName, started); err != nil {
		return buildNum, fmt.Errorf("failed to write started: %v", err)
	}
	return buildNum, nil
}

// FinishBuild uploads build artifacts and a finished record.
func (t *TestGrid) FinishBuild(ctx context.Context, buildNum int, finished *testgrid.Finished, dir string) error {
	// archive test artifacts
	err := t.writeArtifactDir(ctx, buildNum, dir)
	if err != nil {
		return fmt.Errorf("couldn't write report results: %v", err)
	}

	// record finished information
	if err = t.writeBuildFile(ctx, buildNum, finishedFileName, finished); err != nil {
		return fmt.Errorf("failed to write finished: %v", err)
	}
	return nil
}
