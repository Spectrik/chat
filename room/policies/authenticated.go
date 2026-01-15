package policies

import (
	"fmt"

	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type AuthenticatedPolicy struct{}

func (p AuthenticatedPolicy) BeforeJoin(view room.RoomView, args protocol.JoinRoomArgs) error {
	if !view.Authed {
		return fmt.Errorf("You must be authenticated to join this room")
	}

	return nil
}
