package authz

import "github.com/ondrej/chat/client"

type Authorizer interface {
	CanCreateRoom(id *client.Client) error
	CanMute(id *client.Client) error
	// later: CanJoinRoom, CanDeleteRoom, CanKick, ...
}
