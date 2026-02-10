package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	velerosubstr = "managed-velero"
	logsBucket   = "osde2e-logs"
)

// Pre-compiled regex for extracting cluster name from bucket name
var clusterNameRegex = regexp.MustCompile(`^(osde2e-[^-]+)-`)

// isS3BucketFromActiveCluster checks if an S3 bucket belongs to an active cluster
// Returns true if the bucket should be skipped (belongs to active cluster), false if it can be cleaned up
func isS3BucketFromActiveCluster(bucketName string, activeClusters map[string]bool) bool {
	// Extract cluster name from bucket name
	// Example: "osde2e-i5u38-image-registry-us-west-2-abcdef" -> "osde2e-i5u38"
	matches := clusterNameRegex.FindStringSubmatch(bucketName)
	if len(matches) >= 2 {
		clusterName := matches[1]
		if activeClusters[clusterName] {
			log.Printf("Skipping S3 bucket for active cluster %s: %s\n", clusterName, bucketName)
			return true
		}
	}
	return false
}

// CreateS3URL creates an S3 URL from a bucket and a key string.
func CreateS3URL(bucket string, keys ...string) string {
	strippedBucket := strings.Trim(bucket, "/")

	strippedKeys := make([]string, len(keys))
	for i, key := range keys {
		strippedKeys[i] = strings.Trim(key, "/")
	}

	s3JoinArray := []string{"s3:/", strippedBucket}
	s3JoinArray = append(s3JoinArray, strippedKeys...)

	return strings.Join(s3JoinArray, "/")
}

// CleanupS3Buckets finds buckets with substring "osde2e-" or "managed-velero",
// then deletes bucket objects and then buckets
// Ignores buckets belonging to active clusters.
func (CcsAwsSession *ccsAwsSession) CleanupS3Buckets(activeClusters map[string]bool, dryrun bool, sendSummary bool,
	deletedCounter *int, failedCounter *int, errorBuilder *strings.Builder,
) error {
	err := CcsAwsSession.GetAWSSessions()
	if err != nil {
		return err
	}

	result, err := CcsAwsSession.s3.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	// Setup BatchDeleteIterator to iterate through a list of objects.
	batchDeleteClient := s3manager.NewBatchDeleteWithClient(CcsAwsSession.s3)

	for _, bucket := range result.Buckets {
		if (strings.Contains(*bucket.Name, rolesubstr) || strings.Contains(*bucket.Name, velerosubstr)) && !isS3BucketFromActiveCluster(*bucket.Name, activeClusters) && *bucket.Name != logsBucket {
			fmt.Printf("Bucket will be deleted: %s\n", bucket)
			if !dryrun {
				iter := s3manager.NewDeleteListIterator(CcsAwsSession.s3, &s3.ListObjectsInput{
					Bucket: bucket.Name,
				})
				if err := batchDeleteClient.Delete(aws.BackgroundContext(), iter); err != nil {
					errorMsg := fmt.Sprintf("error deleting objects from bucket %s, skipping: %s", *bucket.Name, err)
					fmt.Println(errorMsg)
					*failedCounter++
					if sendSummary && errorBuilder.Len() < 10000 {
						errorBuilder.WriteString(strings.ReplaceAll(errorMsg, `""`, ""))
					}
					continue
				}
				fmt.Println("Deleted object(s) from bucket")
				if _, err := CcsAwsSession.s3.DeleteBucket(&s3.DeleteBucketInput{
					Bucket: bucket.Name,
				}); err != nil {
					errorMsg := fmt.Sprintf("error deleting bucket: %s: %s", *bucket.Name, err)
					fmt.Println(errorMsg)
					*failedCounter++
					if sendSummary && errorBuilder.Len() < config.SlackMessageLength {
						errorBuilder.WriteString(strings.ReplaceAll(errorMsg, `""`, ""))
					}
					continue
				}
				fmt.Println("Deleted bucket")
				*deletedCounter++
			}
		}
	}

	return nil
}

// =============================================================================
// Test artifact uploader
// =============================================================================

// S3Uploader handles uploading test artifacts to S3.
type S3Uploader struct {
	s3Client  *s3.S3              // cached S3 client for presigned URLs
	uploader  *s3manager.Uploader // cached uploader for batch uploads
	bucket    string
	component string // component name for organizing artifacts (e.g., "osd-example-operator")
	category  string // top-level category for organizing artifacts (e.g., "test-results")
	urlExpiry time.Duration
}

// S3UploadResult contains information about uploaded files.
type S3UploadResult struct {
	S3URI        string
	PresignedURL string
	Key          string
	Size         int64
}

// NewS3Uploader creates a new S3 uploader instance using configuration from viper.
// Upload is automatically enabled when LOG_BUCKET is set.
// Reuses the global CcsAwsSession for AWS credentials and session management.
// The component parameter is used to organize artifacts in S3 (e.g., "osd-example-operator").
//
// TODO: Refactor to use dependency injection instead of viper globals.
// Should accept (bucket, region, component string) parameters for better testability
// and reusability. Caller should check config and decide whether to upload.
func NewS3Uploader(component string) (*S3Uploader, error) {
	bucket := viper.GetString(config.Tests.LogBucket)

	// Ensure region is set (default to us-east-1)
	if viper.GetString(config.AWSRegion) == "" {
		viper.Set(config.AWSRegion, "us-east-1")
	}

	// Use the global AWS session infrastructure
	sess, err := CcsAwsSession.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS session: %w", err)
	}

	if component == "" {
		component = "unknown"
	}

	return &S3Uploader{
		s3Client:  s3.New(sess),
		uploader:  s3manager.NewUploader(sess),
		bucket:    bucket,
		component: component,
		category:  "test-results",  // fixed category for S3 path organization
		urlExpiry: 168 * time.Hour, // 7 days (max for IAM user credentials)
	}, nil
}

// BuildS3Key constructs the S3 key path for organizing artifacts.
// Format: <category>/<component>/<date>/<job-id>/
// Example: test-results/osd-example-operator/2026-01-24/abc123/
func (u *S3Uploader) BuildS3Key() string {
	date := time.Now().UTC().Format("2006-01-02")

	jobID := viper.GetString(config.JobID)
	if jobID == "" || jobID == "-1" {
		jobID = viper.GetString(config.Suffix)
	}
	if jobID == "" {
		jobID = fmt.Sprintf("run-%d", time.Now().Unix())
	}

	return path.Join(u.category, u.component, date, jobID)
}

// artifactExtensions maps file extensions to Content-Type headers for S3 uploads.
// Explicit values ensure consistent browser behavior across platforms (presigned URLs).
var artifactExtensions = map[string]string{
	".xml":  "text/xml; charset=utf-8",
	".log":  "text/plain; charset=utf-8",
	".yaml": "text/plain; charset=utf-8",
	".yml":  "text/plain; charset=utf-8",
	".json": "application/json",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".csv":  "text/csv; charset=utf-8",
	".txt":  "text/plain; charset=utf-8",
	".html": "text/html; charset=utf-8",
}

var allowedFilenames = map[string]bool{
	"test_output":   true,
	"summary":       true,
	"osde2e-full":   true,
	"cluster-state": true,
}

func contentTypeForFile(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ct, ok := artifactExtensions[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

func shouldUploadFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := artifactExtensions[ext]; ok {
		return true
	}
	baseName := strings.ToLower(strings.TrimSuffix(filepath.Base(filename), ext))
	return allowedFilenames[baseName]
}

// UploadDirectory uploads files matching allowed extensions to S3.
func (u *S3Uploader) UploadDirectory(srcDir string) ([]S3UploadResult, error) {
	if u == nil {
		return nil, nil
	}

	baseKey := u.BuildS3Key()
	var results []S3UploadResult
	var skippedCount int

	log.Printf("Starting S3 upload from %s to %s", srcDir, CreateS3URL(u.bucket, baseKey))

	err := filepath.WalkDir(srcDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(relPath), ".") {
			return nil
		}

		if !shouldUploadFile(filePath) {
			skippedCount++
			return nil
		}

		s3Key := path.Join(baseKey, relPath)

		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Warning: failed to open %s: %v", filePath, err)
			return nil
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			log.Printf("Warning: failed to stat %s: %v", filePath, err)
			return nil
		}

		contentType := contentTypeForFile(filePath)

		_, err = u.uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(u.bucket),
			Key:         aws.String(s3Key),
			Body:        file,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			log.Printf("Warning: failed to upload %s: %v", filePath, err)
			return nil // Continue with other files; partial upload is better than none
		}

		presignedURL, err := u.generatePresignedURL(s3Key)
		if err != nil {
			log.Printf("Warning: failed to generate presigned URL for %s: %v", s3Key, err)
			presignedURL = ""
		}

		results = append(results, S3UploadResult{
			S3URI:        CreateS3URL(u.bucket, s3Key),
			PresignedURL: presignedURL,
			Key:          s3Key,
			Size:         fileInfo.Size(),
		})

		log.Printf("Uploaded: %s (%d bytes)", relPath, fileInfo.Size())
		return nil
	})
	if err != nil {
		return results, fmt.Errorf("error walking directory: %w", err)
	}

	log.Printf("S3 upload complete: %d files uploaded, %d files skipped", len(results), skippedCount)
	return results, nil
}

func (u *S3Uploader) generatePresignedURL(key string) (string, error) {
	req, _ := u.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return req.Presign(u.urlExpiry)
}

// LogS3UploadSummary prints upload summary and writes artifact URLs for downstream systems.
func LogS3UploadSummary(results []S3UploadResult) {
	if len(results) == 0 {
		log.Println("No files were uploaded to S3")
		return
	}

	log.Println("=== S3 Upload Summary ===")
	log.Printf("Uploaded %d files", len(results))

	var totalSize int64
	for _, r := range results {
		totalSize += r.Size
	}
	log.Printf("Total size: %d bytes", totalSize)

	log.Println("\n=== Presigned URLs (valid for 7 days) ===")
	for _, r := range results {
		if strings.HasSuffix(r.Key, ".xml") || strings.HasSuffix(r.Key, ".log") {
			log.Printf("%s:\n  %s", filepath.Base(r.Key), r.PresignedURL)
		}
	}

	if len(results) > 0 {
		baseKey := path.Dir(results[0].Key)
		log.Printf("\nAll artifacts: %s", CreateS3URL(viper.GetString(config.Tests.LogBucket), baseKey))
	}

	writeArtifactsJSON(results)
}

// ArtifactsJSON is written to stdout and /dev/termination-log for downstream consumption.
// Only key URLs are included due to k8s termination message 4KB limit.
type ArtifactsJSON struct {
	S3URI    string `json:"s3Uri"`
	JUnitURL string `json:"junitUrl,omitempty"`
	LogsURL  string `json:"logsUrl,omitempty"`
}

func writeArtifactsJSON(results []S3UploadResult) {
	if len(results) == 0 {
		return
	}

	// Extract base path: category/component/date/job-id
	var baseKey string
	parts := strings.Split(results[0].Key, "/")
	if len(parts) >= 4 {
		baseKey = strings.Join(parts[:4], "/")
	} else {
		baseKey = path.Dir(results[0].Key)
	}

	artifacts := ArtifactsJSON{
		S3URI: CreateS3URL(viper.GetString(config.Tests.LogBucket), baseKey),
	}

	for _, r := range results {
		baseName := filepath.Base(r.Key)
		if strings.HasPrefix(baseName, "junit") && strings.HasSuffix(baseName, ".xml") && artifacts.JUnitURL == "" {
			artifacts.JUnitURL = r.PresignedURL
		}
		if (baseName == "test_output.log" || baseName == "osde2e-full.log") && artifacts.LogsURL == "" {
			artifacts.LogsURL = r.PresignedURL
		}
	}

	// SetEscapeHTML(false) keeps presigned URLs valid (prevents & -> \u0026)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(artifacts); err != nil {
		log.Printf("Warning: failed to marshal artifacts JSON: %v", err)
		return
	}
	data := bytes.TrimSpace(buf.Bytes())

	fmt.Printf("\n###OSDE2E_ARTIFACTS_JSON###\n%s\n###END_ARTIFACTS_JSON###\n", string(data))
	writeTerminationMessage(data)
}

// writeTerminationMessage writes to /dev/termination-log (k8s job result pattern).
func writeTerminationMessage(data []byte) {
	if err := os.WriteFile("/dev/termination-log", data, 0o644); err != nil {
		log.Printf("Note: Could not write termination message: %v", err)
	}
}
