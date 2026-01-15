package protocol

import "github.com/ondrej/chat/client"

type SayArgs struct {
	Room    string
	Text    string
	Sayer   *client.Client
}
