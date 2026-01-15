package room

import (
	"time"

	"github.com/ondrej/chat/client"
)

type RoomCtx struct {
	State    *RoomState
	Policies PolicySet
	Now      func() time.Time
}

func (rc *RoomCtx) View(c *client.Client) RoomView {
    return RoomView{
        Room:        rc.State.Name,
        ClientCount: len(rc.State.Clients),
        Muted:       rc.State.IsMuted(c),
        Who:         c.ClientName(),
        Authed:      c.Authenticated(),
        Identity:    c.Identity,
    }
}

type Command interface {
	Apply(*RoomCtx) ([]Event, error)
	Name() string
}
