package room

import (
	"github.com/ondrej/chat/protocol"
)

type JoinPolicy interface {
	BeforeJoin(view RoomView, args protocol.JoinRoomArgs) error
}

type SayPolicy interface {
	BeforeSay(view RoomView, text string) error
}

type PolicySet struct {
	Join []JoinPolicy
	Say  []SayPolicy
}
