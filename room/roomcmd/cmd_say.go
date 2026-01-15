package roomcmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/room"
)

var ErrMuted = errors.New("You are muted")

type SayCmd struct {
	Client   *client.Client
	Text     string
}

func (SayCmd) Name() string { return "say" }

func (c SayCmd) Apply(rc *room.RoomCtx) ([]room.Event, error) {
	s:= rc.State

	if !s.IsMember(c.Client) {
		return nil, fmt.Errorf("You are not in a room: %s", s.Name)
	}

	if s.IsMuted(c.Client) {
		return []room.Event{
			room.Direct{To: c.Client, Text: ErrMuted.Error(), At: time.Now()},
		}, ErrMuted
	}

	view := rc.View(c.Client)
	for _, p := range rc.Policies.Say {
		if err := p.BeforeSay(view, c.Text); err != nil {
			return []room.Event{
				room.Direct{To: c.Client, Text: err.Error(), At: time.Now()},
			}, err
		}
	}

	now := time.Now()
	recipients := make([]*client.Client, 0, len(rc.State.Clients))
	for c := range rc.State.Clients { recipients = append(recipients, c) }

	return []room.Event{
		room.Broadcast{
			Room: s.Name,
			Text: c.Text,
			At:   now,
			Recipients: recipients,
		},
	}, nil
}
