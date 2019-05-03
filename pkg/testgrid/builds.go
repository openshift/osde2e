package testgrid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	testgrid "k8s.io/test-infra/testgrid/metadata"
)

const (
	startedFileName  = "started.json"
	finishedFileName = "finished.json"
	artifactsDir     = "artifacts"
)

func (t *TestGrid) writeStarted(ctx context.Context, buildNum int, timestamp int64) error {
	started := &testgrid.Started{
		Timestamp: timestamp,
	}

	data, err := json.Marshal(started)
	if err != nil {
		return fmt.Errorf("failed encoding started file: %v", err)
	}

	return t.writeBuildFile(ctx, buildNum, startedFileName, data)
}

func (t *TestGrid) writeFinished(ctx context.Context, buildNum int, finished testgrid.Finished) error {
	data, err := json.Marshal(&finished)
	if err != nil {
		return fmt.Errorf("failed encoding finished file: %v", err)
	}

	return t.writeBuildFile(ctx, buildNum, finishedFileName, data)
}

func (t *TestGrid) getBuildFile(ctx context.Context, buildNum int, filename string) ([]byte, error) {
	key := t.buildFileKey(buildNum, filename)
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

func (t *TestGrid) writeBuildFile(ctx context.Context, buildNum int, filename string, data []byte) error {
	key := t.buildFileKey(buildNum, filename)
	w := t.bucket.Object(key).NewWriter(ctx)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed while writing file '%s' for build %d: %v", key, buildNum, err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to finish writing file '%s' for build %d: %v", key, buildNum, err)
	}
	return nil
}

func (t *TestGrid) buildFileKey(buildNum int, filename string) string {
	return filepath.Join(t.prefix, strconv.Itoa(buildNum), filename)
}

func (t *TestGrid) writeReportDir(ctx context.Context, buildNum int, dir string) error {
	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fInfo := range dirInfo {
		// only upload files
		if !fInfo.IsDir() {
			fileName := filepath.Join(dir, fInfo.Name())

			f, err := os.Open(fileName)
			if err != nil {
				return fmt.Errorf("error opening '%s': %v", fileName, err)
			}

			name := filepath.Join(artifactsDir, fInfo.Name())
			key := t.buildFileKey(buildNum, name)
			w := t.bucket.Object(key).NewWriter(ctx)
			if _, err = io.Copy(w, f); err != nil {
				return fmt.Errorf("error uploading '%s' as '%s': %v", fileName, key, err)
			} else if err = w.Close(); err != nil {
				return fmt.Errorf("error finishing upload of '%s' as '%s': %v", fileName, key, err)
			} else if err = f.Close(); err != nil {
				log.Printf("Error closing file '%s': %v", fileName, err)
			}
		}
	}
	return nil
}
