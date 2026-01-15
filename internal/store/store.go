package store

import "github.com/ondrej/chat/internal/user"

type UserStore interface {
	GetUserByUsername(username string) (*user.User, error)
}
