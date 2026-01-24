package aws

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
)

const (
	velerosubstr = "managed-velero"
	logsBucket   = "osde2e-logs"
)

// isS3BucketFromActiveCluster checks if an S3 bucket belongs to an active cluster
// Returns true if the bucket should be skipped (belongs to active cluster), false if it can be cleaned up
func isS3BucketFromActiveCluster(bucketName string, activeClusters map[string]bool) bool {
	// Extract cluster name from bucket name
	// Example: "osde2e-i5u38-image-registry-us-west-2-abcdef" -> "osde2e-i5u38"
	re := regexp.MustCompile(`^(osde2e-[^-]+)-`)
	matches := re.FindStringSubmatch(bucketName)
	if len(matches) >= 2 {
		clusterName := matches[1]
		if activeClusters[clusterName] {
			log.Printf("Skipping S3 bucket for active cluster %s: %s\n", clusterName, bucketName)
			return true
		}
	}
	return false
}

// ReadFromS3Session reads a key from S3 using given AWS context.
func ReadFromS3Session(session *session.Session, inputKey string) ([]byte, error) {
	bucket, key, err := ParseS3URL(inputKey)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse S3 URL: %v", err)
	}

	downloader := s3manager.NewDownloader(session)

	buffer := aws.NewWriteAtBuffer([]byte{})

	_, err = downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// WriteToS3Session writes the given byte array to S3.
func WriteToS3Session(session *session.Session, outputKey string, data []byte) {
	bucket, key, err := ParseS3URL(outputKey)
	if err != nil {
		log.Printf("error trying to parse S3 URL %s: %v", outputKey, err)
		return
	}

	uploader := s3manager.NewUploader(session)

	reader := bytes.NewReader(data)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		log.Printf("Failed to upload to s3 %s", err.Error())
		return
	}
	log.Printf("Uploaded to %s", outputKey)
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

// ParseS3URL parses an S3 url into a bucket and key.
func ParseS3URL(s3URL string) (string, string, error) {
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		return "", "", err
	}

	return parsedURL.Host, parsedURL.Path, nil
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
	session   *session.Session
	s3Client  *s3.S3
	bucket    string
	region    string
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
func NewS3Uploader() (*S3Uploader, error) {
	bucket := viper.GetString(config.Tests.LogBucket)
	if bucket == "" {
		// S3 upload disabled - no bucket configured
		return nil, nil
	}

	// Use AWS_REGION if set, otherwise default to us-east-1 (where osde2e-loki-logs bucket is)
	region := viper.GetString(config.AWSRegion)
	if region == "" {
		region = "us-east-1"
	}

	// Create AWS session with explicit region
	awsAccessKey := viper.GetString(config.AWSAccessKey)
	awsSecretAccessKey := viper.GetString(config.AWSSecretAccessKey)
	awsProfile := viper.GetString(config.AWSProfile)

	options := session.Options{
		Config: aws.Config{
			Region: aws.String(region),
		},
	}

	if awsProfile != "" {
		options.Profile = awsProfile
	} else if awsAccessKey != "" && awsSecretAccessKey != "" {
		options.Config.Credentials = credentials.NewStaticCredentials(awsAccessKey, awsSecretAccessKey, "")
	}

	sess, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Uploader{
		session:   sess,
		s3Client:  s3.New(sess),
		bucket:    bucket,
		region:    region,
		category:  "test-results",  // fixed category for S3 path organization
		urlExpiry: 168 * time.Hour, // 7 days (max for IAM user credentials)
	}, nil
}

// BuildS3Key constructs the S3 key path for organizing artifacts.
// Format: <category>/<component>/<date>/<job-id>/
// Example: test-results/osd-example-operator/2026-01-24/abc123/
func (u *S3Uploader) BuildS3Key() string {
	component := deriveComponent()

	date := time.Now().UTC().Format("2006-01-02")

	jobID := viper.GetString(config.JobID)
	if jobID == "" || jobID == "-1" {
		jobID = viper.GetString(config.Suffix)
	}
	if jobID == "" {
		jobID = fmt.Sprintf("run-%d", time.Now().Unix())
	}

	parts := []string{}
	if u.category != "" {
		parts = append(parts, strings.Trim(u.category, "/"))
	}
	parts = append(parts, component, date, jobID)

	return strings.Join(parts, "/")
}

// deriveComponent determines the component name from the test image.
// It extracts a meaningful name from the test image path to organize S3 artifacts.
// Examples:
//
//	quay.io/org/osd-example-operator-e2e:tag -> osd-example-operator
//	quay.io/org/my-service-test:latest -> my-service
func deriveComponent() string {
	testSuites, err := config.GetTestSuites()
	if err == nil && len(testSuites) > 0 {
		imageName := testSuites[0].Image
		if component := extractNameFromImage(imageName); component != "" {
			log.Printf("Derived component from test image: %s -> %s", imageName, component)
			return component
		}
	}

	log.Println("Could not derive component, using fallback: unknown")
	return "unknown"
}

// extractNameFromImage extracts a meaningful name from a container image path.
// It strips the registry, organization, tag, and common test suffixes.
// Examples:
//
//	quay.io/org/osd-example-operator-e2e:tag -> osd-example-operator
//	quay.io/org/my-service-test:latest -> my-service
//	quay.io/org/simple:v1 -> simple
func extractNameFromImage(image string) string {
	if image == "" {
		return ""
	}

	// Remove tag (everything after :)
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		image = image[:idx]
	}

	// Remove registry and org (everything before last /)
	if idx := strings.LastIndex(image, "/"); idx != -1 {
		image = image[idx+1:]
	}

	// Strip common test suffixes
	suffixes := []string{"-e2e", "-test", "-tests", "-harness"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(image, suffix) {
			image = strings.TrimSuffix(image, suffix)
			break
		}
	}

	return image
}

// UploadDirectory uploads all files from a directory to S3.
// Returns a list of upload results with S3 URIs and presigned URLs.
func (u *S3Uploader) UploadDirectory(srcDir string) ([]S3UploadResult, error) {
	if u == nil {
		return nil, nil
	}

	baseKey := u.BuildS3Key()
	var results []S3UploadResult
	uploader := s3manager.NewUploader(u.session)

	log.Printf("Starting S3 upload from %s to s3://%s/%s/", srcDir, u.bucket, baseKey)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip hidden files and marker files
		if strings.HasPrefix(filepath.Base(relPath), ".") {
			return nil
		}

		// Construct S3 key
		s3Key := fmt.Sprintf("%s/%s", baseKey, relPath)

		// Read file
		file, err := os.Open(path)
		if err != nil {
			log.Printf("Warning: failed to open file %s: %v", path, err)
			return nil // Continue with other files
		}
		defer file.Close()

		// Get file info for size
		fileInfo, err := file.Stat()
		if err != nil {
			log.Printf("Warning: failed to stat file %s: %v", path, err)
			return nil
		}

		// Determine content type
		contentType := "application/octet-stream"
		switch {
		case strings.HasSuffix(path, ".xml"):
			contentType = "application/xml"
		case strings.HasSuffix(path, ".json"):
			contentType = "application/json"
		case strings.HasSuffix(path, ".log"), strings.HasSuffix(path, ".txt"):
			contentType = "text/plain"
		case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
			contentType = "text/yaml"
		case strings.HasSuffix(path, ".html"):
			contentType = "text/html"
		}

		// Upload file
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(u.bucket),
			Key:         aws.String(s3Key),
			Body:        file,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			log.Printf("Warning: failed to upload %s: %v", path, err)
			return nil // Continue with other files
		}

		// Generate presigned URL
		presignedURL, err := u.generatePresignedURL(s3Key)
		if err != nil {
			log.Printf("Warning: failed to generate presigned URL for %s: %v", s3Key, err)
			presignedURL = ""
		}

		s3URI := fmt.Sprintf("s3://%s/%s", u.bucket, s3Key)
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

	// Group by directory for cleaner output
	var totalSize int64
	for _, r := range results {
		totalSize += r.Size
	}
	log.Printf("Total size: %d bytes", totalSize)

	// Print presigned URLs for key files (JUnit XML, logs)
	log.Println("\n=== Presigned URLs (valid for 7 days) ===")
	for _, r := range results {
		if strings.HasSuffix(r.Key, ".xml") || strings.HasSuffix(r.Key, ".log") || strings.HasSuffix(r.Key, "test_output.log") {
			log.Printf("%s:\n  %s", filepath.Base(r.Key), r.PresignedURL)
		}
	}

	// Print base S3 URI
	if len(results) > 0 {
		baseKey := filepath.Dir(results[0].Key)
		log.Printf("\nAll artifacts: s3://%s/%s/", viper.GetString(config.Tests.LogBucket), baseKey)
	}
}

// IsS3UploadEnabled returns whether S3 upload is enabled.
// Upload is enabled when LOG_BUCKET is configured.
func IsS3UploadEnabled() bool {
	return viper.GetString(config.Tests.LogBucket) != ""
}
