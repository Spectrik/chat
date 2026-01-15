package roomcmd

import (
	"fmt"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/room"
)

type LeaveCmd struct {
	Client   *client.Client
}

func (LeaveCmd) Name() string { return "leave" }

func (c LeaveCmd) Apply(rc *room.RoomCtx) ([]room.Event, error) {
	s := rc.State

	if _, ok := s.Clients[c.Client]; !ok {
		return nil, nil
	}

	delete(s.Clients, c.Client)

	now := rc.Now()
	name := c.Client.ClientName()
	recipients := make([]*client.Client, 0, len(rc.State.Clients))
	for c := range rc.State.Clients { recipients = append(recipients, c) }

	return []room.Event{
		room.Left{Room: s.Name, Who: c.Client, At: now},
		room.Broadcast{Room: s.Name, Text: name + " left.", At: now, Recipients: recipients},
		room.Direct{To: c.Client, Text: fmt.Sprintf("Left room: %s", s.Name), At: now},
	}, nil
}
