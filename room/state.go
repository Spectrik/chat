package room

import (
	"time"

	"github.com/ondrej/chat/client"
)

type RoomState struct {
	Name      string
	CreatedAt time.Time

	Clients map[*client.Client]struct{}
	Mutes   map[client.Identity]Mute
}

func newState(name string) RoomState {
	return RoomState{
		Name:      name,
		CreatedAt: time.Now(),
		Clients:   make(map[*client.Client]struct{}),
		Mutes:     make(map[client.Identity]Mute),
	}
}

func (s *RoomState) IsMember(c *client.Client) bool {
	_, ok := s.Clients[c]
	return ok
}

func (s *RoomState) IsMuted(c *client.Client) bool {
	// reuse tvou logiku (User/IP)
	auth := c.Authenticated()
	for key := range s.Mutes {
		if auth && key.User != nil && c.Identity.User != nil && key.User.Equal(*c.Identity.User) {
			return true
		}
		if key.Ip != "" && c.Identity.Ip == key.Ip {
			return true
		}
	}
	return false
}
