package filestore

import (
	"errors"

	"github.com/ondrej/chat/internal/authn"
	"github.com/ondrej/chat/internal/user"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type FileAuthenticator struct {
	store *FileStore
}

func NewService(store *FileStore) *FileAuthenticator {
	return &FileAuthenticator{store: store}
}

func (s *FileAuthenticator) Authenticate(username, password string) (*user.User, error) {

	rec, err := s.store.getRecord(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if rec.Disabled {
		return nil, ErrInvalidCredentials
	}

	if !authn.CheckPasswordHash(password, rec.Password) {
		return  nil, ErrInvalidCredentials
	}

	return recordToDomain(rec), nil
}
