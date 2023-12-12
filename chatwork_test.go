package testkit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatwork(t *testing.T) {
	harness := New(t, Providers(
		&EnvProvider{},
	))

	chatworkRoom := harness.ChatworkRoom(t)

	t.Logf("chatworkRoom: %+v", chatworkRoom)

	c := &Chatwork{
		Token: chatworkRoom.Token,
	}

	_, err := c.GetMessages(chatworkRoom.ID)
	require.NoError(t, err)

	r, err := c.PostMessage(chatworkRoom.ID, "hello world")
	require.NoError(t, err)

	t.Logf("r: %+v", r)

	ms, err := c.GetMessages(chatworkRoom.ID)
	require.NoError(t, err)

	if len(ms) != 1 {
		t.Fatalf("unexpected number of messages: %d", len(ms))
	}

	if ms[0].Body != "hello world" {
		t.Fatalf("unexpected message body: %s", ms[0].Body)
	}
}
