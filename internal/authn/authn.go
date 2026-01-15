package authn

import "github.com/ondrej/chat/internal/user"

type Authenticator interface {
	Authenticate(username, password string) (*user.User, error)
}
