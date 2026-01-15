package protocol

import "github.com/ondrej/chat/client"

type DirectMessageArgs struct {
	User    string
	Text    string
	Sayer   *client.Client
}
