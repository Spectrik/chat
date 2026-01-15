package roommanager

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ondrej/chat/client"
	"github.com/ondrej/chat/room"
)

var log = slog.Default().With("component", "roommanager")

type RoomManager struct {
	Rooms map[string]*room.Room
	MemberOf map[*client.Client]map[string]struct{}
	ops chan func()
	cleanup  chan struct{}
	stop  chan struct{}
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*room.Room),
		cleanup:  make(chan struct{}),
		stop:  make(chan struct{}),
		ops:   make(chan func(), 128),
		MemberOf: make(map[*client.Client]map[string]struct{}),
	}
}

func (rm *RoomManager) ListRoomsOf(c *client.Client) []string {
    var out []string
    rm.DoSync(func() {
        m := rm.MemberOf[c]
        if m == nil { return }
        out = make([]string, 0, len(m))
        for name := range m { out = append(out, name) }
    })

    return out
}

func (rm *RoomManager) GetRoom(name string) (*room.Room, bool) {
	var exists bool
	var r *room.Room

	rm.DoSync(func() {
		r, exists = rm.Rooms[name]
	})

	return r, exists
}

func (rm *RoomManager) RoomList() []string {
	var list []string

	rm.DoSync(func() {
		list = make([]string, 0, len(rm.Rooms))
		for name := range rm.Rooms {
			list = append(list, name)
		}
	})

	return list
}

func (rm *RoomManager) CreateRoom(name string, policies room.PolicySet) (*room.Room, error) {
	var err error
	var r *room.Room

	rm.DoSync(func() {
		if _, ok := rm.Rooms[name]; ok {
			err = fmt.Errorf("Room %s already exists!", name)
			return
		}

		r = room.NewRoom(name, policies)
		rm.Rooms[name] = r
	})

	return r, err
}

func (rm *RoomManager) Run() {
	for {
		select {
		case op := <-rm.ops:
			op()
		case <-rm.cleanup:
			rm.roomCleanup()
		case <-rm.stop:
			for _, room := range rm.Rooms {
				room.Stop()
			}
			return
		}
	}
}

func (rm *RoomManager) DoSync(op func()) {
	done := make(chan struct{})
	rm.ops <- func() {
		op()
		close(done)
	}
	<-done
}

func (rm *RoomManager) roomCleanup() {
	log.Debug("Running room cleanup")

	for name, r := range rm.Rooms {
		count, _ := r.ClientCount(context.TODO()) // TODO
		if count == 0 && r.Permanent != true {
			log.Debug("Cleaning up room:", name)
			r.Stop()
			delete(rm.Rooms, name)
		}
	}
}

func (rm *RoomManager) StartJanitor(interval time.Duration) func() {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
				case <-t.C:
					select {
						case rm.cleanup <- struct{}{}:
						default:
					}
				case <-done:
					return
			}
		}
	}()

	return func() { close(done) }
}

// TODO
// func (rm *RoomManager) LoadRoomsFromConfig(cfg chatconfig.ChatConfig) {
// 	for _, rc := range cfg.Rooms {
// 		rm.CreateRoom(rc.Name)
// 	}
// }
