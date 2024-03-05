package gitops_slack_bot

import "net"

type Bot struct {
	LastError error
}

// Starts a slack bot that listens to the provided listener and calls the onReceveMessage function
// when a message is received.
// The onReceveMessage function should return an error if the bot should stop.
// The bot does not return the error from onReceveMessage.
// Instead, it sets the LastError field to the error and stops the bot,
// so that the caller can check the LastError field to see the result of the onReceveMessage function.
//
// The bot triggers the onReceveMessage function when the triggerMessage is received.
func Start(ln net.Listener, triggerMessage string, onReceveMessage func(string) error) (*Bot, error) {
	bot := &Bot{}
	return bot, nil
}
