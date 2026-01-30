package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"mime"
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
						errorBuilder.WriteString(strings.Replace(errorMsg, `""`, "", -1))
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
						errorBuilder.WriteString(strings.Replace(errorMsg, `""`, "", -1))
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
func NewS3Uploader(component string) (*S3Uploader, error) {
	bucket := viper.GetString(config.Tests.LogBucket)
	if bucket == "" {
		// S3 upload disabled - no bucket configured
		return nil, nil
	}

	// Ensure region is set (default to us-east-1 for osde2e-loki-logs bucket)
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

	// path.Join handles empty strings correctly and always uses forward slashes
	return path.Join(u.category, u.component, date, jobID)
}

// UploadDirectory uploads all files from a directory to S3.
// Returns a list of upload results with S3 URIs and presigned URLs.
func (u *S3Uploader) UploadDirectory(srcDir string) ([]S3UploadResult, error) {
	if u == nil {
		return nil, nil
	}

	baseKey := u.BuildS3Key()
	var results []S3UploadResult

	log.Printf("Starting S3 upload from %s to %s", srcDir, CreateS3URL(u.bucket, baseKey))

	err := filepath.WalkDir(srcDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip hidden files and marker files
		if strings.HasPrefix(filepath.Base(relPath), ".") {
			return nil
		}

		// Construct S3 key using path.Join (always uses forward slashes, correct for S3)
		s3Key := path.Join(baseKey, relPath)

		// Read file
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("Warning: failed to open file %s: %v", filePath, err)
			return nil // Continue with other files
		}
		defer file.Close()

		// Get file info for size
		fileInfo, err := file.Stat()
		if err != nil {
			log.Printf("Warning: failed to stat file %s: %v", filePath, err)
			return nil
		}

		// Determine content type using standard mime package
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Upload file using cached uploader for better performance
		_, err = u.uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(u.bucket),
			Key:         aws.String(s3Key),
			Body:        file,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			log.Printf("Warning: failed to upload %s: %v", filePath, err)
			return nil // Continue with other files
		}

		// Generate presigned URL
		presignedURL, err := u.generatePresignedURL(s3Key)
		if err != nil {
			log.Printf("Warning: failed to generate presigned URL for %s: %v", s3Key, err)
			presignedURL = ""
		}

		// Reuse existing CreateS3URL helper
		s3URI := CreateS3URL(u.bucket, s3Key)
		results = append(results, S3UploadResult{
			S3URI:        s3URI,
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

	return results, nil
}

// generatePresignedURL creates a presigned URL for accessing an S3 object.
func (u *S3Uploader) generatePresignedURL(key string) (string, error) {
	req, _ := u.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return req.Presign(u.urlExpiry)
}

// LogS3UploadSummary prints a summary of uploaded files with their access URLs.
func LogS3UploadSummary(results []S3UploadResult) {
	if len(results) == 0 {
		log.Println("No files were uploaded to S3")
		return
	}

	log.Println("=== S3 Upload Summary ===")
	log.Printf("Uploaded %d files", len(results))

	// Calculate total size
	var totalSize int64
	for _, r := range results {
		totalSize += r.Size
	}
	log.Printf("Total size: %d bytes", totalSize)

	// Print presigned URLs for key files (JUnit XML, logs)
	log.Println("\n=== Presigned URLs (valid for 7 days) ===")
	for _, r := range results {
		// .log suffix covers test_output.log, no need for separate check
		if strings.HasSuffix(r.Key, ".xml") || strings.HasSuffix(r.Key, ".log") {
			log.Printf("%s:\n  %s", filepath.Base(r.Key), r.PresignedURL)
		}
	}

	// Print base S3 URI
	if len(results) > 0 {
		baseKey := path.Dir(results[0].Key)
		log.Printf("\nAll artifacts: %s", CreateS3URL(viper.GetString(config.Tests.LogBucket), baseKey))
	}

	// Write structured JSON for downstream systems (qontract-reconcile, etc.)
	writeArtifactsJSON(results)
}

// ArtifactsJSON is the structured output format for S3 artifact URLs.
// This JSON is written to stdout and termination message for downstream consumption.
// Note: Kubernetes termination message is limited to 4KB, so we only include key URLs.
type ArtifactsJSON struct {
	S3URI    string `json:"s3Uri"`
	JUnitURL string `json:"junitUrl,omitempty"`
	LogsURL  string `json:"logsUrl,omitempty"`
}

// writeArtifactsJSON outputs artifact URLs in a well-known JSON format.
// This enables downstream systems to parse and link to artifacts.
func writeArtifactsJSON(results []S3UploadResult) {
	if len(results) == 0 {
		return
	}

	artifacts := ArtifactsJSON{}

	// Set base S3 URI (use the common prefix path, not the subdirectory)
	// Example: if key is "test-results/component/2026-01-30/run-123/install/junit.xml"
	// we want "test-results/component/2026-01-30/run-123"
	baseKey := ""
	if len(results) > 0 {
		// Extract base path from the first result key
		// Assume structure: <category>/<component>/<date>/<job-id>/<optional-subdir>/<file>
		parts := strings.Split(results[0].Key, "/")
		if len(parts) >= 4 {
			// Take first 4 parts: category/component/date/job-id
			baseKey = strings.Join(parts[:4], "/")
		} else {
			// Fallback: use directory of first file
			baseKey = path.Dir(results[0].Key)
		}
	}
	artifacts.S3URI = CreateS3URL(viper.GetString(config.Tests.LogBucket), baseKey)

	// Find key artifact URLs (prioritize junit XML and main log file)
	for _, r := range results {
		// Match JUnit XML files (junit*.xml pattern)
		baseName := filepath.Base(r.Key)
		if strings.HasPrefix(baseName, "junit") && strings.HasSuffix(baseName, ".xml") && artifacts.JUnitURL == "" {
			artifacts.JUnitURL = r.PresignedURL
		}
		// Match main log files
		if (baseName == "test_output.log" || baseName == "osde2e-full.log") && artifacts.LogsURL == "" {
			artifacts.LogsURL = r.PresignedURL
		}
	}

	// Use encoder with SetEscapeHTML(false) to prevent & from being escaped as \u0026
	// This ensures presigned URLs remain valid and clickable
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(artifacts); err != nil {
		log.Printf("Warning: failed to marshal artifacts JSON: %v", err)
		return
	}
	data := bytes.TrimSpace(buf.Bytes())

	// Output with marker for easy parsing
	fmt.Printf("\n###OSDE2E_ARTIFACTS_JSON###\n%s\n###END_ARTIFACTS_JSON###\n", string(data))

	// Write to termination message (standard k8s pattern for job results)
	// Note: termination message is limited to 4KB, keeping only essential URLs
	writeTerminationMessage(data)
}

// writeTerminationMessage writes to /dev/termination-log for Kubernetes job results.
// This is the standard pattern for making job output available in pod status.
func writeTerminationMessage(data []byte) {
	terminationPath := "/dev/termination-log"
	if err := os.WriteFile(terminationPath, data, 0o644); err != nil {
		// Not an error - termination-log may not exist outside k8s
		log.Printf("Note: Could not write termination message: %v", err)
	}
}
