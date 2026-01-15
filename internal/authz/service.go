package authz

import (
	"errors"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/internal/user"
)

var ErrNotAuthorized = errors.New("Not authorized to perform this action")
var ErrNotAuthenticated = errors.New("User not authenticated")

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CanCreateRoom(id *client.Client) error {
	if ! id.Authenticated() {
		return ErrNotAuthenticated
	}

	if id.Identity.User.Role < user.RoleAdmin {
		return ErrNotAuthorized
	}

	return nil
}

func (s *Service) CanMute(id *client.Client) error {
	if !id.Authenticated() {
		return ErrNotAuthorized
	}

	if id.Identity.User.Role < user.RoleMod {
		return ErrNotAuthorized
	}

	return nil
}
