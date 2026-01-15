package clientregistry

import (
	"errors"
	"fmt"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/internal/store"
)

var ErrNotFound = errors.New("User not found")
type Registry struct {
	ops chan func()
	byUserID map[string]*client.Client
	registered	store.UserStore
}

var ReservedNames = map[string]struct{} {
	"admin": {},
}

func NewRegistry(store store.UserStore) *Registry{
	r := &Registry{
		ops:       make(chan func(), 256),
		byUserID: make(map[string]*client.Client),
		registered: store,
	}

	go r.loop()
	return r
}

func (r *Registry) loop() {
	for op := range r.ops {
		op()
	}
}

func (r *Registry) Do(fn func()) {
	done := make(chan struct{})
	r.ops <- func() { fn(); close(done) }
	<-done
}

func (r *Registry) Register(c *client.Client) {
	r.Do(func() {
		r.byUserID[c.ClientName()] = c
	})
}

func (r *Registry) Unregister(c *client.Client) {
	r.Do(func() {
		delete(r.byUserID, c.ClientName())
	})
}

func (r *Registry) isUsernameTaken(username string) bool {
	var exists bool = false

	r.Do(func() {
		_, exists = r.byUserID[username]
	})

	return exists
}

func (r *Registry) WithClientByName(username string, fn func(*client.Client) error) error {
    var err error

    r.Do(func() {
        c, ok := r.byUserID[username]
        if !ok {
            err = ErrNotFound
            return
        }
        err = fn(c)
    })

    return err
}

// func (r *Registry) GetClientByName(username string) (*client.Client, bool) {
// 	var c *client.Client
// 	var ok bool
// 	r.Do(func() {
// 		c, ok = r.byUserID[username]
// 	})

// 	return c, ok
// }

func (r *Registry) ValidateUsername(username string) error {
	if username == "" {
        return fmt.Errorf("Username can not be empty")
	}

	if r.isUsernameTaken(username) {
		return fmt.Errorf(username + " is already taken")
		// TODO generate random one
	}

	if _, ok := ReservedNames[username]; ok {
		return fmt.Errorf(username + " is a reserved username")
	}

	// TODO: conflict s perzistent userem, optional rename
	// query do storage
	u, _ := r.registered.GetUserByUsername(username)

	if u != nil {
		return fmt.Errorf(username + " is a registered username")
	}

	return nil
}
