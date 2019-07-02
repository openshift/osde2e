package testgrid

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
)

func (t *TestGrid) writeArtifactDir(ctx context.Context, buildNum int, dir string) error {
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

			// check if data is compressed
			gzipped := false
			if strings.HasSuffix(key, ".gzip") {
				gzipped = true
				key = strings.TrimSuffix(key, ".gzip")
			}

			obj := t.bucket.Object(key)
			w := obj.NewWriter(ctx)
			if _, err = io.Copy(w, f); err != nil {
				return fmt.Errorf("error uploading '%s' as '%s': %v", fileName, key, err)
			} else if err = w.Close(); err != nil {
				return fmt.Errorf("error finishing upload of '%s' as '%s': %v", fileName, key, err)
			} else if err = f.Close(); err != nil {
				log.Printf("Error closing file '%s': %v", fileName, err)
			}

			// update metadata if data is compressed
			if gzipped {
				attrs := storage.ObjectAttrsToUpdate{
					ContentEncoding: "gzip",
				}

				if strings.HasSuffix(key, ".json") {
					attrs.ContentType = "application/json"
				}

				if _, err = obj.Update(ctx, attrs); err != nil {
					return fmt.Errorf("couldn't update metadata with gzip info: %v", err)
				}
			}
		}
	}
	return nil
}
