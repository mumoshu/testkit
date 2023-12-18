package testkit_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mumoshu/testkit"
	"github.com/stretchr/testify/require"
)

func TestTestKit(t *testing.T) {
	harness := testkit.New(t, testkit.Providers(
		&testkit.KubectlProvider{},
		&testkit.EKSCTLProvider{},
		&testkit.TerraformProvider{},
		&testkit.EnvProvider{},
	))

	eksCluster := harness.EKSCluster(t)
	s3Bucket := harness.S3Bucket(t)
	ns := harness.KubernetesNamespace(t, testkit.KubeconfigPath(eksCluster.KubeconfigPath))

	awsConfig := s3Bucket.AWSV2Config(t)

	//
	// Examples of your application code.
	//

	appErr := runYourApp(yourAppConfig{
		awsConfig:    awsConfig,
		s3BucketName: s3Bucket.Name,
	})
	if appErr != nil {
		t.Fatal(appErr)
	}

	//
	// Examples of your assertion code using testkit.
	//

	s3BucketClient := testkit.NewS3BucketClient(awsConfig, s3Bucket.Name)
	require.Equal(t,
		"hello world",
		s3BucketClient.GetString(t, "my-object"),
	)

	kubectl := testkit.NewKubectl(eksCluster.KubeconfigPath)
	require.Contains(t,
		kubectl.Capture(t, "get", "ns", ns.Name),
		ns.Name,
	)
}

type yourAppConfig struct {
	awsConfig    aws.Config
	s3BucketName string
}

func runYourApp(c yourAppConfig) error {
	ctx := context.Background()

	awsS3Svc := s3.NewFromConfig(c.awsConfig)

	putObjParams := &s3.PutObjectInput{
		Bucket: &c.s3BucketName,
		Key:    aws.String("my-object"),
		Body:   bytes.NewReader([]byte("hello world")),
	}
	_, err := awsS3Svc.PutObject(ctx, putObjParams)

	return err
}

type testProvider struct {
	setup bool
}

func (p *testProvider) Setup() error {
	p.setup = true
	return nil
}

func (p *testProvider) Cleanup() error {
	p.setup = false
	return nil
}

func TestNew(t *testing.T) {
	p := &testProvider{}
	cnt := 0

	t.Run("setup", func(t *testing.T) {
		_ = testkit.New(t, testkit.Providers(p))

		require.True(t, p.setup)
		cnt++
	})

	// See if the provider is cleaned up.
	require.False(t, p.setup)
	// Ensure that the setup function is called only once.
	require.Equal(t, 1, cnt)
}

func TestKindKubectl(t *testing.T) {
	os.Unsetenv("KUBECONFIG")

	tk := testkit.New(t, testkit.Providers(&testkit.KindProvider{}))
	kc := tk.KubernetesCluster(t)

	harness := testkit.New(t, testkit.Providers(&testkit.KubectlProvider{
		DefaultKubeconfigPath: kc.KubeconfigPath,
	}))

	ns := harness.KubernetesNamespace(t)
	ns2 := harness.KubernetesNamespace(t)
	defaultCM1 := harness.KubernetesConfigMap(t)
	defaultCM2 := harness.KubernetesConfigMap(t)
	nsCM1 := harness.KubernetesConfigMap(t, testkit.KubernetesConfigMapNamespace(ns.Name))
	nsCM2 := harness.KubernetesConfigMap(t, testkit.KubernetesConfigMapNamespace(ns.Name))

	require.Equal(t, ns, ns2)
	require.Equal(t, defaultCM1, defaultCM2)
	require.Equal(t, nsCM1, nsCM2)
}

func TestTerraform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	os.Unsetenv("KUBECONFIG")

	vpcID := os.Getenv("TESTKIT_VPC_ID")
	if vpcID == "" {
		t.Error("TESTKIT_VPC_ID environment variable is not set")
		t.FailNow()
	}

	tk := testkit.New(t, testkit.Providers(&testkit.TerraformProvider{
		WorkspacePath: "testdata/terraform",
		Vars: map[string]string{
			"prefix": "testkit-",
			"region": "ap-northeast-1",
			"vpc_id": vpcID,
		},
	}), testkit.RetainResourcesOnFailure())
	bucket := tk.S3Bucket(t)
	eksCluster := tk.EKSCluster(t)

	awsConfig := bucket.AWSV2Config(t)
	s3BucketClient := testkit.NewS3BucketClient(awsConfig, bucket.Name)
	s3BucketClient.PutString(t, "my-object", "hello world")
	require.Equal(t,
		"hello world",
		s3BucketClient.GetString(t, "my-object"),
	)
	// Otherwise, the test cleanup i.e. the bucket deletion attempt by
	// terraform-destroy will fail because the bucket is not empty.
	s3BucketClient.Delete(t, "my-object")

	kubectl := testkit.NewKubectl(eksCluster.KubeconfigPath)
	require.Contains(t,
		kubectl.Capture(t, "get", "ns", "default"),
		"default",
	)
}

func TestTerraformS3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tk := testkit.New(t, testkit.Providers(&testkit.TerraformProvider{
		WorkspacePath: "testdata/terraform-s3",
		Vars: map[string]string{
			"prefix": "testkit-tfs3-",
		},
	}), testkit.RetainResourcesOnFailure())
	bucket := tk.S3Bucket(t)

	awsConfig := bucket.AWSV2Config(t)
	s3BucketClient := testkit.NewS3BucketClient(awsConfig, bucket.Name)

	require.Empty(t,
		s3BucketClient.ListKeys(t, ""),
	)

	s3BucketClient.PutString(t, "my-object", "hello world")
	require.Equal(t,
		"hello world",
		s3BucketClient.GetString(t, "my-object"),
	)

	require.Equal(t,
		[]string{"my-object"},
		s3BucketClient.ListKeys(t, "my-"),
	)

	require.Empty(t,
		s3BucketClient.ListKeys(t, "your-"),
	)

	// Otherwise, the test cleanup i.e. the bucket deletion attempt by
	// terraform-destroy will fail because the bucket is not empty.
	s3BucketClient.Delete(t, "my-object")

	require.Empty(t,
		s3BucketClient.ListKeys(t, ""),
	)
}
