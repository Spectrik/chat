package app

import (
	"fmt"

	"github.com/ondrej/chat/room"
)

func pumpRoomEvents(r *room.Room) {
	for ev := range r.Events() {
		switch e := ev.(type) {
		case room.Broadcast:
			for _, c := range e.Recipients {
				c.Send(fmt.Sprintf("%s [%s] [%s] %s", e.At, e.Room, c.ClientName(), e.Text))
			}
		case room.Direct:
			e.To.Send(e.Text)
		}
	}
}
