package roomcmd

import (
	"fmt"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/protocol"
	"github.com/ondrej/chat/room"
)

type JoinCmd struct {
	Args     protocol.JoinRoomArgs
}

func (JoinCmd) Name() string { return "join" }

func (c JoinCmd) Apply(rc *room.RoomCtx) ([]room.Event, error) {
	view := rc.View(c.Args.Client)

	for _, p := range rc.Policies.Join {
		if err := p.BeforeJoin(view, c.Args); err != nil {
			return nil, err
		}
	}

	rc.State.Clients[c.Args.Client] = struct{}{}

	now := rc.Now()
	recipients := make([]*client.Client, 0, len(rc.State.Clients))
	for c := range rc.State.Clients { recipients = append(recipients, c) }

	return []room.Event{
		room.Joined{Room: rc.State.Name, Who: c.Args.Client, At: now},
		room.Broadcast{Room: rc.State.Name, Text: fmt.Sprintf("%s has joined", c.Args.Client.ClientName()), At: now, Recipients: recipients},
	}, nil
}
