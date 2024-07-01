package testkit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mumoshu/testkit"
	gitops_slack_bot "github.com/mumoshu/testkit/testapps/gitops-slack-bot"
	"github.com/stretchr/testify/require"
)

func TestTestKit(t *testing.T) {
	harness := testkit.New(t, testkit.Providers(
		&testkit.KubectlProvider{},
		&testkit.EKSCTLProvider{},
		&testkit.TerraformProvider{
			WorkspacePath: "testdata/terraform",
		},
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

	os.Setenv("TESTKIT_LOG", "debug")

	tk := testkit.New(t, testkit.Providers(&testkit.KindProvider{}))
	// KindProvider creates a new Kubernetes cluster using kind
	// on first call to KubernetesCluster.
	// Setting TESTKIT_LOG=debug will show the kubeconfig path and
	// content in the test output, for debugging purposes.
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

	k := testkit.NewKubernetes(kc.KubeconfigPath)

	testkit.PollUntil(t, func() bool {
		return len(k.ListReadyNodeNames(t)) == 1
	}, 20*time.Second)

	helm := testkit.NewHelm(kc.KubeconfigPath)
	helm.UpgradeOrInstall(t, "my-release", "testdata/helm-chart")
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

// TestGitOpsSlackBot tests the example Slack bot that listens to /deploy commands
// in the Slack channel to deploy something to the GitHub repository.
//
// The bot is implemented in the `testapps/gitops-slack-bot` directory.
// The bot is written in golang and run within this test, not as a separate service in a container or a cluster.
//
// The bot is tested by sending a /deploy command to the Slack channel,
// and then checking if the bot has created and merged a pull request with expected content against
// the GitHub repository.
//
// The bot is connected with Slack by exposting the local HTTP endpoint served by the bot to the internet
// via ngrok.
// In a real-world scenario, the bot would be deployed to e.g. a Kubernetes cluster and exposed to the internet
// via a LoadBalancer service. That can be done using our Terraform provider.
func TestGitOpsSlackBot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	h := testkit.New(t, testkit.Providers(
		&testkit.GitHubWritableRepositoriesEnvProvider{},
		&testkit.EnvProvider{},
	), testkit.RetainResourcesOnFailure())

	repo := h.GitHubWritableRepository(t)

	ngrokConf := testkit.NgrokConfigFromEnv()

	// Starts ngrok to expose the local HTTP endpoint to the internet
	ln, err := testkit.ListenNgrok(t, ngrokConf)
	require.NoError(t, err)

	// Note that you need to register the ngrok's endpoint URL
	// to the Slack incoming webhook configuration.

	// We write the current time of the precision of a second to the repository via Slack
	// so that we can verity that the bot has actually triggered by the /deploy
	// message and written the content to the repository.
	data := time.Now().Format(time.RFC3339)
	commitMsg := "Deployed by testkit slack bot at " + data

	triggerMessage := "/deploy"

	bot, err := gitops_slack_bot.Start(ln, triggerMessage, func(message string) error {
		return repo.WriteFileE("testkit.test", data, commitMsg)
	})
	require.NoError(t, err)

	// And we presume that you've already registered the ngrok's external URL
	// to the Slack incoming webhook configuration.
	// Do also note that the URL is the one that is provided to the EnvProvider above
	// and contained the h.SlackChannel return value below, which is used by PostMessage.

	slackCh := h.SlackChannel(t)
	slackCh.SendMessage(t, triggerMessage)

	time.Sleep(5 * time.Second)

	require.NoError(t, bot.LastError)

	commits := repo.FindCommits(t, "main", "")
	require.NotEmpty(t, commits)
	require.Len(t, commits, 1)

	commit := commits[0]

	require.Equal(t, commitMsg, commit.Message)
}

// One-time test to verify the Slack challenge request to register the bot's endpoint URL.
//
// Set approprite env variables and run this once immediately after
// starting the challenge on Slack.
func TestSlackChallenge(t *testing.T) {
	if os.Getenv("DO_SLACK_CHALLENGE") == "" {
		t.Skip("Set DO_SLACK_CHALLENGE=1 to run this test")
	}

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ngrokConf := testkit.NgrokConfigFromEnv()

	// Starts ngrok to expose the local HTTP endpoint to the internet
	ln, err := testkit.ListenNgrok(t, ngrokConf)
	require.NoError(t, err)

	// Runs a http server that responds to Slack's challenge request
	// to verify the endpoint URL.
	// The server is expected to be exposed to the internet via ngrok.

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	type challengeRequest struct {
		Token     string `json:"token"`
		Challenge string `json:"challenge"`
		Type      string `json:"type"`
	}

	var (
		errCh = make(chan error, 1)

		server = &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req challengeRequest

				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(req.Challenge))

				cancel()
			}),
		}
	)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	go func() {
		errCh <- server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatalf("server.Serve: %v", ctx.Err())

		// Wait for the server to shutdown gracefully
		time.Sleep(5 * time.Second)
	}
}
