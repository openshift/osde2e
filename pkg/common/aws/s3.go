package aws

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/openshift/osde2e/pkg/common/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	velerosubstr = "managed-velero"
	logsBucket   = "osde2e-logs"
)

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
	return
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
// older than given duration, then deletes bucket objects and then buckets
func (CcsAwsSession *ccsAwsSession) CleanupS3Buckets(olderthan time.Duration, dryrun bool, sendSummary bool,
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
		if (strings.Contains(*bucket.Name, rolesubstr) || strings.Contains(*bucket.Name, velerosubstr)) && *bucket.Name != logsBucket && time.Since(*bucket.CreationDate) > olderthan {
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
