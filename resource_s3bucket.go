package testkit

import "testing"

// S3BucketProvider is a provider that can provision an S3 bucket.
// Any provider that can create an S3 bucket should implement this interface.
type S3BucketProvider interface {
	GetS3Bucket(opts ...S3BucketOption) (*S3Bucket, error)
}

type S3Bucket struct {
	Region, profile, Name string
}

type S3BucketConfig struct {
	ID string
}

type S3BucketOption func(*S3BucketConfig)

func (tk *TestKit) S3Bucket(t *testing.T, opts ...S3BucketOption) *S3Bucket {
	t.Helper()

	var cp S3BucketProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(S3BucketProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no S3BucketProvider found")
	}

	s3Bucket, err := cp.GetS3Bucket(opts...)
	if err != nil {
		t.Fatalf("unable to get s3 bucket: %v", err)
	}

	return s3Bucket
}
