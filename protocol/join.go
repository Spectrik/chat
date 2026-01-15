package protocol

import "github.com/ondrej/chat/client"

type JoinRoomArgs struct {
	Password string
	Room     string
	Client   *client.Client
}
