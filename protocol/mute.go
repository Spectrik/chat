package protocol

import "github.com/ondrej/chat/client"

type MuteArgs struct {
	Duration int
	Room    string
	Target   string
	Actor    *client.Client
}
