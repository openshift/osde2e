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

// NewTestGrid configures a new TestGrid.
func NewTestGrid(bucket, prefix string, b64ServiceAccount []byte) (*TestGrid, error) {
	if bucket == "" {
		return nil, errors.New("bucket for TestGrid is not set")
	} else if b64ServiceAccount == nil || len(b64ServiceAccount) == 0 {
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

// StartBuild uploads a `testgrid.Started` record and updates the latest build to point to it.
func (t *TestGrid) StartBuild(ctx context.Context, timestamp int64) (buildNum int, err error) {
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
	if err = t.writeStarted(ctx, buildNum, timestamp); err != nil {
		return buildNum, fmt.Errorf("failed to write started: %v", err)
	}
	return buildNum, nil
}

// FinishBuild uploads build artifacts and a `testgrid.Finished` record.
func (t *TestGrid) FinishBuild(ctx context.Context, buildNum int, result testgrid.Finished, dir string) error {
	err := t.writeReportDir(ctx, buildNum, dir)
	if err != nil {
		return fmt.Errorf("couldn't write report results: %v", err)
	}

	if err = t.writeFinished(ctx, buildNum, result); err != nil {
		return fmt.Errorf("failed to write finished: %v", err)
	}
	return nil
}
