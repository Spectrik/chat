package room

import "github.com/ondrej/chat/client"

type RoomView struct {
	Room        string
	ClientCount int
	Muted       bool
	Who         string
	Authed      bool
	Identity    client.Identity
}
