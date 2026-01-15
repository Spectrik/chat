package policies

import (
	"fmt"

	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type PasswordPolicy struct {
	Password string
}

func (p PasswordPolicy) BeforeJoin(view room.RoomView, args protocol.JoinRoomArgs) error {
	if p.Password == "" {
		return nil
	}

	if args.Password != p.Password {
		return fmt.Errorf("invalid password")
	}

	return nil
}
