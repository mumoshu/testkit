package testkit

import (
	"log"
	"os"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

type SlackMessaging interface {
	SendMessage(t *testing.T, message string)
}

type SlackMessagingClient struct {
	slackClient *slack.Client

	channel string
}

func NewSlackMessagingClient(token, channel string) *SlackMessagingClient {
	return &SlackMessagingClient{
		slackClient: slack.New(
			token, slack.OptionLog(
				log.New(os.Stdout, "testkit-slack: ", log.Lshortfile|log.LstdFlags),
			),
		),
		channel: channel,
	}
}

func (c *SlackMessagingClient) SendMessage(t *testing.T, message string) {
	t.Helper()

	_, _, err := c.slackClient.PostMessage(
		c.channel,
		slack.MsgOptionText(message, false),
	)
	// If you get "missing_scope" erorr here, you need to add the "chat:write" scope to your bot token.
	// Click "Add an OAuth Scope" in the "OAuth & Permissions" page of your app settings,
	// add the "chat:write" scope, and reinstall the app to your workspace.
	//
	// If you get "invalid_auth" error here, you need to reinstall the app to your workspace.
	// Go to the "OAuth & Permissions" page of your app settings, and click "Reinstall App".
	//
	// If you get "channel_not_found" error here, you need to invite your bot to the channel.
	// Go to the channel, and type "/add" and select "Add apps to this channel",
	// and select your bot app.
	if err != nil {
		t.Logf("unable to send message to channel %q: %v", c.channel, err)
	}
	require.NoError(t, err)
}
