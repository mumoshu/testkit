package testkit

import "testing"

type ChatworkRoomProvider interface {
	GetChatworkRoom(opts ...ChatworkRoomOption) (*ChatworkRoom, error)
}

type ChatworkRoom struct {
	ID    string
	Token string
}

type ChatworkRoomConfig struct {
	ID string
}

type ChatworkRoomOption func(*ChatworkRoomConfig)

func (tk *TestKit) ChatworkRoom(t *testing.T, opts ...ChatworkRoomOption) *ChatworkRoom {
	t.Helper()

	var cp ChatworkRoomProvider
	for _, p := range tk.availableProviders {
		var ok bool

		cp, ok = p.(ChatworkRoomProvider)
		if ok {
			break
		}
	}

	if cp == nil {
		t.Fatal("no ChatworkRoomProvider found")
	}

	r, err := cp.GetChatworkRoom(opts...)
	if err != nil {
		t.Fatalf("unable to get Chatwork room: %v", err)
	}

	return r
}
