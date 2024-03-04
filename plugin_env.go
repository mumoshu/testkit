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

func (p *EnvProvider) GetGitHubRepository(_ ...GitHubRepositoryOption) (*GitHubRepository, error) {
	repo := os.Getenv("TESTKIT_GITHUB_REPO")
	if repo == "" {
		return nil, fmt.Errorf("TESTKIT_GITHUB_REPO environment variable is not set")
	}

	token := os.Getenv("TESTKIT_GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TESTKIT_GITHUB_TOKEN environment variable is not set")
	}

	return &GitHubRepository{
		Name:  repo,
		Token: token,
	}, nil
}

func (p *EnvProvider) GetSlackChannel(_ ...SlackChannelOption) (*SlackChannel, error) {
	channel := os.Getenv("TESTKIT_SLACK_CHANNEL")
	if channel == "" {
		return nil, fmt.Errorf("TESTKIT_SLACK_CHANNEL environment variable is not set")
	}

	botToken := os.Getenv("TESTKIT_SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TESTKIT_SLACK_BOT_TOKEN environment variable is not set")
	}

	appToken := os.Getenv("TESTKIT_SLACK_APP_TOKEN")
	if appToken == "" {
		return nil, fmt.Errorf("TESTKIT_SLACK_APP_TOKEN environment variable is not set")
	}

	inURL := os.Getenv("TESTKIT_SLACK_INCOMING_WEBHOOK_URL")
	if inURL == "" {
		return nil, fmt.Errorf("TESTKIT_SLACK_INCOMING_WEBHOOK_URL environment variable is not set")
	}

	return &SlackChannel{
		ID:                 channel,
		BotToken:           botToken,
		AppToken:           appToken,
		IncomingWebhookURL: inURL,
	}, nil
}

func (p *EnvProvider) GetChatworkRoom(_ ...ChatworkRoomOption) (*ChatworkRoom, error) {
	room := os.Getenv("TESTKIT_CHATWORK_ROOM_ID")
	if room == "" {
		return nil, fmt.Errorf("TESTKIT_CHATWORK_ROOM environment variable is not set")
	}

	token := os.Getenv("TESTKIT_CHATWORK_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TESTKIT_CHATWORK_TOKEN environment variable is not set")
	}

	return &ChatworkRoom{
		ID:    room,
		Token: token,
	}, nil
}
