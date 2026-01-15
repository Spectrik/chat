package policies

import (
	"fmt"

	"github.com/ondrej/chat/internal/user"
	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type isAdminPolicy struct{}

func (p isAdminPolicy) BeforeJoin(view room.RoomView, args protocol.JoinRoomArgs) error {
	if view.Identity.User.Role != user.RoleAdmin {
		return fmt.Errorf("Authorization denied: User %s does not have an admin role", view.Identity.User.Username)
	}

	return nil
}
