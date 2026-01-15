package policies

import (
	"fmt"

	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type CapacityPolicy struct {
	Max int
}

func (p CapacityPolicy) BeforeJoin(view room.RoomView, args protocol.JoinRoomArgs) error {
	if p.Max <= 0 {
		return nil
	}

	if view.ClientCount >= p.Max {
		return fmt.Errorf("Room is full (max %d)", p.Max)
	}

	return nil
}
