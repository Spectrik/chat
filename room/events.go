package room

import (
	"time"

	"github.com/ondrej/chat/client"
)

type Event interface{ event() }

type Broadcast struct {
	Room string
	Text string
	At   time.Time
	Recipients []*client.Client
}

func (Broadcast) event() {}

type Direct struct {
	To   *client.Client
	Text string
	At   time.Time
}
func (Direct) event() {}

type Joined struct {
	Room string
	Who  *client.Client
	At   time.Time
}
func (Joined) event() {}

type Left struct {
	Room string
	Who  *client.Client
	At   time.Time
}
func (Left) event() {}