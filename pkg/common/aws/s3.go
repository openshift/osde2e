package aws

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// ReadFromS3 reads a key from S3 using the global AWS context.
func ReadFromS3(inputKey string) ([]byte, error) {
	bucket, key, err := ParseS3URL(inputKey)
	if err != nil {
		return nil, fmt.Errorf("error trying to parse S3 URL: %v", err)
	}

	session, err := AWSSession.getSession()
	if err != nil {
		return nil, err
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

// WriteToS3 writes the given byte array to S3.
func WriteToS3(outputKey string, data []byte) error {
	bucket, key, err := ParseS3URL(outputKey)
	if err != nil {
		return fmt.Errorf("error trying to parse S3 URL: %v", err)
	}

	session, err := AWSSession.getSession()
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(session)

	reader := bytes.NewReader(data)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})

	if err != nil {
		return err
	}

	log.Printf("Uploaded to %s", outputKey)

	return nil
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
