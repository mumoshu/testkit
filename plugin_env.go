package testkit

import (
	"fmt"
	"os"
	"strings"
)

type EnvProvider struct {
}

var _ Provider = &EnvProvider{}
var _ S3BucketProvider = &EnvProvider{}
var _ EKSClusterProvider = &EnvProvider{}

func (p *EnvProvider) Setup() error {
	var found bool
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "TESTKIT_") {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no TESTKIT_* environment variables found")
	}

	return nil
}

func (p *EnvProvider) Cleanup() error {
	return nil
}

func (p *EnvProvider) GetEKSCluster(_ ...EKSClusterOption) (*EKSCluster, error) {
	kubeconfigPath := os.Getenv("TESTKIT_KUBECONFIG")
	if kubeconfigPath == "" {
		return nil, fmt.Errorf("TESTKIT_KUBECONFIG environment variable is not set")
	}

	return &EKSCluster{
		KubeconfigPath: kubeconfigPath,
	}, nil
}

func (p *EnvProvider) GetS3Bucket(_ ...S3BucketOption) (*S3Bucket, error) {
	bucketName := os.Getenv("TESTKIT_S3_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("TESTKIT_S3_BUCKET_NAME environment variable is not set")
	}

	region := os.Getenv("TESTKIT_S3_BUCKET_REGION")
	if region == "" {
		return nil, fmt.Errorf("TESTKIT_S3_BUCKET_REGION environment variable is not set")
	}

	profile := os.Getenv("TESTKIT_S3_BUCKET_PROFILE")

	return &S3Bucket{
		Name:    bucketName,
		Region:  region,
		profile: profile,
	}, nil
}
