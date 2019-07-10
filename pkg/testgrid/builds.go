package testgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	testgrid "k8s.io/test-infra/testgrid/metadata"
)

const (
	startedFileName  = "started.json"
	finishedFileName = "finished.json"
	artifactsDir     = "artifacts"
)

// Started retrieves information for buildNum that was created when it started.
func (t *TestGrid) Started(ctx context.Context, buildNum int) (started testgrid.Started, err error) {
	data, err := t.getBuildFile(ctx, buildNum, startedFileName)
	if err != nil {
		return started, fmt.Errorf("failed retrieving started record for build %d: %v", buildNum, err)
	}

	if err = json.Unmarshal(data, &started); err != nil {
		err = fmt.Errorf("failed decoding started record for build %d: %v", buildNum, err)
	}
	return
}

// Finished retrieves results for buildNum that were created when it finished running.
func (t *TestGrid) Finished(ctx context.Context, buildNum int) (finished testgrid.Finished, err error) {
	data, err := t.getBuildFile(ctx, buildNum, finishedFileName)
	if err != nil {
		return finished, fmt.Errorf("failed retrieving started record for build %d: %v", buildNum, err)
	}

	if err = json.Unmarshal(data, &finished); err != nil {
		err = fmt.Errorf("failed decoding started record for build %d: %v", buildNum, err)
	}
	return
}

func (t *TestGrid) getBuildFile(ctx context.Context, buildNum int, filename ...string) ([]byte, error) {
	key := t.buildFileKey(buildNum, filename...)
	rdr, err := t.bucket.Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to request file '%s' for build %d: %v", key, buildNum, err)
	}
	defer rdr.Close()

	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer file '%s' for build %d: %v", key, buildNum, err)
	}
	return data, nil
}

func (t *TestGrid) writeBuildFile(ctx context.Context, buildNum int, filename string, out interface{}) (err error) {
	key := t.buildFileKey(buildNum, filename)
	var data []byte

	// marshal out if necessary
	switch typedOut := out.(type) {
	case []byte:
		data = typedOut
	default:
		if data, err = json.Marshal(out); err != nil {
			return fmt.Errorf("failed encoding file '%s' for build %d: %v", filename, buildNum, err)
		}
	}

	// write file to gcs
	w := t.bucket.Object(key).NewWriter(ctx)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed while writing file '%s' for build %d: %v", key, buildNum, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to finish writing file '%s' for build %d: %v", key, buildNum, err)
	}
	return nil
}

func (t *TestGrid) buildFileKey(buildNum int, filenames ...string) string {
	buildPath := append([]string{t.prefix, strconv.Itoa(buildNum)}, filenames...)
	return filepath.Join(buildPath...)
}
