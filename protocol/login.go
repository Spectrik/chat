package protocol

import "github.com/ondrej/chat/client"

type LoginArgs struct {
	Username string
	Password string
	Client   *client.Client
}
