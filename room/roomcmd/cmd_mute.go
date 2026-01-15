package roomcmd

import (
	"fmt"
	"time"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/room"
)

type MuteCmd struct {
	Actor    *client.Client
	Target   string
	Duration time.Duration
}

func (MuteCmd) Name() string { return "mute" }

func (c MuteCmd) Apply(rc *room.RoomCtx) ([]room.Event, error) {
	var target *client.Client
	s := rc.State

	for cl := range s.Clients {
		if cl.ClientName() == c.Target {
			target = cl
			break
		}
	}

	now := time.Now()
	if target == nil {
		return []room.Event{
			room.Direct{To: c.Actor, Text: "User not present in the target room!", At: now},
		}, nil
	}

	s.Mutes[target.Identity] = room.Mute{When: now, Duration: c.Duration}

	return []room.Event{
		room.Direct{To: c.Actor, Text: fmt.Sprintf("Muted: %s for %.0f min", target.ClientName(), c.Duration.Minutes()), At: now},
	}, nil
}
