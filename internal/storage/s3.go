package storage

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3 struct {
	client *s3.S3
	bucket string
}

func NewS3(bucket, region, endpoint, accessKey, secretKey string) (*S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	return &S3{
		client: s3.New(sess),
		bucket: bucket,
	}, nil
}

func (s *S3) Store(filename string, data io.Reader) error {
	// Convert io.Reader to bytes.Reader to ensure proper seeking behavior for S3
	var body io.ReadSeeker
	if buf, ok := data.(*bytes.Buffer); ok {
		// If it's a bytes.Buffer, create a bytes.Reader from its contents
		body = bytes.NewReader(buf.Bytes())
	} else {
		// For other readers, read all data and create bytes.Reader
		allData, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		body = bytes.NewReader(allData)
	}

	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
		Body:   body,
	})
	return err
}
