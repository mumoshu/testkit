package testkit

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// GetLatestString returns the latest object in the bucket with the given prefix.
// The returned string is the content of the object.
// "Latest" means the object with the latest LastModified timestamp.
func (c *S3BucketClient) GetLatestString(t *testing.T, prefix string) string {
	t.Helper()

	listObjInput := &s3.ListObjectsV2Input{
		Bucket: &c.Bucket,
		Prefix: aws.String(prefix),
	}
	o, err := c.S3Svc.ListObjectsV2(context.Background(), listObjInput)
	if err != nil {
		t.Fatal(err)
	}

	if len(o.Contents) == 0 {
		t.Fatalf("no object found with prefix %q", prefix)
	}

	var latestObj *s3types.Object
	for _, obj := range o.Contents {
		o := obj
		if latestObj == nil {
			latestObj = &o
			continue
		}

		if latestObj.LastModified == nil {
			latestObj = &o
			continue
		}

		if o.LastModified == nil {
			continue
		}

		if latestObj.LastModified.Before(*o.LastModified) {
			latestObj = &o
		}
	}

	getObjInput := &s3.GetObjectInput{
		Bucket: &c.Bucket,
		Key:    latestObj.Key,
	}
	obj, err := c.S3Svc.GetObject(context.Background(), getObjInput)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(obj.Body)
	require.NoError(t, err)

	return string(got)
}

// ListKeys returns the keys of the objects in the bucket with the given prefix.
func (c *S3BucketClient) ListKeys(t *testing.T, prefix string) []string {
	t.Helper()

	listObjInput := &s3.ListObjectsV2Input{
		Bucket: &c.Bucket,
		Prefix: aws.String(prefix),
	}
	o, err := c.S3Svc.ListObjectsV2(context.Background(), listObjInput)
	if err != nil {
		t.Fatal(err)
	}

	var keys []string
	for _, obj := range o.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys
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
