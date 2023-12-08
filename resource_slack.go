package testkit

import "testing"

type SlackChannelProvider interface {
	GetSlackChannel(opts ...SlackChannelOption) (*SlackChannel, error)
}

type SlackChannel struct {
	ID                 string
	Token              string
	IncomingWebhookURL string
}

type SlackChannelConfig struct {
	ID string
}

type SlackChannelOption func(*SlackChannelConfig)

func (tk *TestKit) SlackChannel(t *testing.T, opts ...SlackChannelOption) *SlackChannel {
	t.Helper()

	var cp SlackChannelProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(SlackChannelProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no SlackChannelProvider found")
	}

	slackCh, err := cp.GetSlackChannel(opts...)
	if err != nil {
		t.Fatalf("unable to get Slack channel: %v", err)
	}

	return slackCh
}
