package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	EnvChatworkRoomID = "TESTKIT_CHATWORK_ROOM_ID"
	EnvChatworkToken  = "TESTKIT_CHATWORK_TOKEN"

	DefaultChatworkEndpoint = "https://api.chatwork.com/v2"

	MessagesPath = "/rooms/%s/messages"
)

type Chatwork struct {
	Endpoint string
	Token    string
}

func (c *Chatwork) GetMessages(roomID string) (ChatworkMessages, error) {
	req, err := c.newHttpRequestWithAuth(http.MethodGet, fmt.Sprintf(MessagesPath, roomID))
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	defer func() {
		_ = res.Body.Close()
	}()

	var messages ChatworkMessages

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if err := json.NewDecoder(bytes.NewReader(d)).Decode(&messages); err != nil {
		var r ChatworkResponse
		if err2 := json.NewDecoder(bytes.NewReader(d)).Decode(&r); err2 != nil {
			return nil, fmt.Errorf("unable to decode response body into chatwork response: %s\nErrors[0]=%v, Errors[1]=%v", string(d), err, err2)
		}

		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("chatwork error: %s", strings.Join(r.Errors, ", "))
		}

		return nil, fmt.Errorf("unable to decode response body into messages: %s", string(d))
	}

	return messages, nil
}

type ChatworkResponse struct {
	Errors []string `json:"errors"`
}

type ChatworkMessages []ChatworkMessage

type ChatworkMessage struct {
	MessageID  string          `json:"message_id"`
	Account    ChatworkAccount `json:"account"`
	Body       string          `json:"body"`
	SendTime   uint64          `json:"send_time"`
	UpdateTime uint64          `json:"update_time"`
}

type ChatworkMessagePostResult struct {
	MessageID string `json:"message_id"`
}

type ChatworkAccount struct {
	AccountID      uint64 `json:"account_id"`
	Name           string `json:"name"`
	AvatarImageURL string `json:"avatar_image_url"`
}

func (c *Chatwork) PostMessage(roomID, message string) (*ChatworkMessagePostResult, error) {
	req, err := c.newHttpRequestWithAuth(http.MethodPost, fmt.Sprintf(MessagesPath, roomID))
	if err != nil {
		return nil, err
	}

	req.Body = io.NopCloser(strings.NewReader(fmt.Sprintf("body=%s", message)))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	var r ChatworkMessagePostResult

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if err := json.NewDecoder(bytes.NewReader(d)).Decode(&r); err != nil {
		var r ChatworkResponse
		if err := json.NewDecoder(bytes.NewReader(d)).Decode(&r); err != nil {
			return nil, fmt.Errorf("unable to decode response body: %s", string(d))
		}

		if len(r.Errors) > 0 {
			return nil, fmt.Errorf("chatwork error: %s", strings.Join(r.Errors, ", "))
		}
		return nil, fmt.Errorf("unable to decode response body into message post result: %s", string(d))
	}

	return &r, nil
}

func (c *Chatwork) getEndpoint() string {
	if c.Endpoint == "" {
		return DefaultChatworkEndpoint
	}

	return c.Endpoint
}

func (c *Chatwork) newHttpRequestWithAuth(method, path string) (*http.Request, error) {
	u := c.getEndpoint() + path
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, err
	}

	// https://developer.chatwork.com/docs/endpoints#%E3%83%AA%E3%82%AF%E3%82%A8%E3%82%B9%E3%83%88
	req.Header.Set("X-ChatworkToken", c.Token)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}
