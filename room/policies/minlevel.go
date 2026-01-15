package policies

import (
	"fmt"

	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type MinLevelPolicy struct{
	Minlevel uint
}

func (m MinLevelPolicy) BeforeJoin(view room.RoomView, args protocol.JoinRoomArgs) error {
	var level uint
	if !view.Authed {
		level = 0
	} else {
		level = uint(view.Identity.User.Role)
	}

	if level < m.Minlevel {
		return fmt.Errorf("Authorization denied: You: %s do not have sufficient permissions to enter the room", view.Who)
	}

	return nil
}
