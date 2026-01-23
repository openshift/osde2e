// Package s3upload provides functionality for uploading test artifacts to S3.
package s3upload

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
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

// Uploader handles uploading test artifacts to S3.
type Uploader struct {
	session   *session.Session
	s3Client  *s3.S3
	bucket    string
	region    string
	prefix    string
	urlExpiry time.Duration
}

// UploadResult contains information about uploaded files.
type UploadResult struct {
	S3URI        string
	PresignedURL string
	Key          string
	Size         int64
}

// NewUploader creates a new S3 uploader instance using configuration from viper.
func NewUploader() (*Uploader, error) {
	if !viper.GetBool(config.S3Upload.Enabled) {
		return nil, nil
	}

	bucket := viper.GetString(config.S3Upload.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required when S3 upload is enabled")
	}

	region := viper.GetString(config.S3Upload.Region)
	if region == "" {
		region = viper.GetString(config.AWSRegion)
	}
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

	expiryHours := viper.GetInt(config.S3Upload.PresignedURLExpiry)
	if expiryHours <= 0 {
		expiryHours = 24
	}

	return &Uploader{
		session:   sess,
		s3Client:  s3.New(sess),
		bucket:    bucket,
		region:    region,
		prefix:    viper.GetString(config.S3Upload.Prefix),
		urlExpiry: time.Duration(expiryHours) * time.Hour,
	}, nil
}

// BuildS3Key constructs the S3 key path for organizing artifacts.
// Format: <prefix>/<operator-name>/<date>/<job-id>/
func (u *Uploader) BuildS3Key() string {
	operatorName := deriveOperatorName()

	date := time.Now().UTC().Format("2006-01-02")

	jobID := viper.GetString(config.JobID)
	if jobID == "" || jobID == "-1" {
		jobID = viper.GetString(config.Suffix)
	}
	if jobID == "" {
		jobID = fmt.Sprintf("run-%d", time.Now().Unix())
	}

	parts := []string{}
	if u.prefix != "" {
		parts = append(parts, strings.Trim(u.prefix, "/"))
	}
	parts = append(parts, operatorName, date, jobID)

	return strings.Join(parts, "/")
}

// deriveOperatorName attempts to determine the operator name from configuration.
// Priority:
//  1. Explicit S3_UPLOAD_OPERATOR_NAME env var
//  2. Derive from TEST_IMAGE or TestSuites
//  3. Fallback to "unknown-operator"
func deriveOperatorName() string {
	// Priority 1: Explicit OPERATOR_NAME env var
	if operatorName := viper.GetString(config.S3Upload.OperatorName); operatorName != "" {
		log.Printf("Using explicit operator name: %s", operatorName)
		return operatorName
	}

	// Priority 2: Derive from TEST_IMAGE or TestSuites
	testSuites, err := config.GetTestSuites()
	if err == nil && len(testSuites) > 0 {
		// Use the first test suite image
		imageName := testSuites[0].Image
		if operatorName := extractOperatorNameFromImage(imageName); operatorName != "" {
			log.Printf("Derived operator name from test image: %s -> %s", imageName, operatorName)
			return operatorName
		}
	}

	// Fallback: unknown-operator
	log.Println("Could not derive operator name, using fallback: unknown-operator")
	return "unknown-operator"
}

// extractOperatorNameFromImage extracts the operator name from a container image path.
// Examples:
//
//	quay.io/org/osd-example-operator-e2e:tag -> osd-example-operator
//	quay.io/org/my-operator-test:latest -> my-operator
//	quay.io/org/simple:v1 -> simple
func extractOperatorNameFromImage(image string) string {
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
func (u *Uploader) UploadDirectory(srcDir string) ([]UploadResult, error) {
	if u == nil {
		return nil, nil
	}

	baseKey := u.BuildS3Key()
	var results []UploadResult
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
		results = append(results, UploadResult{
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
func (u *Uploader) generatePresignedURL(key string) (string, error) {
	req, _ := u.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	return req.Presign(u.urlExpiry)
}

// LogUploadSummary prints a summary of uploaded files with their access URLs.
func LogUploadSummary(results []UploadResult) {
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
		log.Printf("\nAll artifacts: s3://%s/%s/", viper.GetString(config.S3Upload.Bucket), baseKey)
	}
}

// IsEnabled returns whether S3 upload is enabled.
func IsEnabled() bool {
	return viper.GetBool(config.S3Upload.Enabled)
}
