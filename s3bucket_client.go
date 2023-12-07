package testkit

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

type S3BucketClient struct {
	S3Svc  *s3.Client
	Bucket string
}

func NewS3BucketClient(awsConfig aws.Config, bucket string) *S3BucketClient {
	s3Svc := s3.NewFromConfig(awsConfig)

	return &S3BucketClient{
		S3Svc:  s3Svc,
		Bucket: bucket,
	}
}

func (c *S3BucketClient) PutString(t *testing.T, key, value string) {
	t.Helper()

	putObjParams := &s3.PutObjectInput{
		Bucket: &c.Bucket,
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(value)),
	}
	_, err := c.S3Svc.PutObject(context.Background(), putObjParams)

	require.NoError(t, err)
}

func (c *S3BucketClient) GetString(t *testing.T, key string) string {
	t.Helper()

	getObjInput := &s3.GetObjectInput{
		Bucket: &c.Bucket,
		Key:    aws.String(key),
	}
	o, err := c.S3Svc.GetObject(context.Background(), getObjInput)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(o.Body)
	require.NoError(t, err)

	return string(got)
}

func (c *S3BucketClient) Delete(t *testing.T, key string) {
	t.Helper()

	deleteObjInput := &s3.DeleteObjectInput{
		Bucket: &c.Bucket,
		Key:    aws.String(key),
	}
	_, err := c.S3Svc.DeleteObject(context.Background(), deleteObjInput)
	require.NoError(t, err)
}
