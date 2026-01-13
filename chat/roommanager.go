package chat

import (
	"fmt"
	"time"

	chatconfig "github.com/ondrej/chat/config"
)

type RoomManager struct {
	rooms map[string]*Room
	memberOf map[*Client]map[string]struct{}
	ops chan func()
	cleanup  chan struct{}
	stop  chan struct{}
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
		cleanup:  make(chan struct{}),
		stop:  make(chan struct{}),
		ops:   make(chan func(), 128),
		memberOf: make(map[*Client]map[string]struct{}),
	}
}

func (rm *RoomManager) RoomList() []string {
	names := make([]string, 0, len(rm.rooms))
	for name := range rm.rooms {
		names = append(names, name)
	}

	return names
}

func (rm *RoomManager) CreateRoom(name string, opts ...RoomOption) *Room {
	room := NewRoom(name, opts...)
	rm.rooms[name] = room
	return room
}

func (rm *RoomManager) Run() {
	for {
		select {
		case op := <-rm.ops:
			op()
		case <-rm.cleanup:
			rm.Cleanup()
		case <-rm.stop:
			for _, room := range rm.rooms {
				room.Stop()
			}
			return
		}
	}
}

func (rm *RoomManager) do(op func()) {
	rm.ops <- op
}

func (rm *RoomManager) Cleanup() {
	for name, room := range rm.rooms {
		if !room.permanent && room.ClientCount() == 0 {
			fmt.Println("Cleaning up room:", name)
			room.Stop()
			delete(rm.rooms, name)
		}
	}
}

func (rm *RoomManager) StartJanitor(interval time.Duration) func() {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(interval)

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

func (rm *RoomManager) JoinRoom(roomName string, c *Client) {
	rm.do(func() {
		r := rm.rooms[roomName]
		if r == nil {
			r = NewRoom(roomName)
			rm.rooms[roomName] = r
		}

		r.Join(c)

		if rm.memberOf[c] == nil {
			rm.memberOf[c] = make(map[string]struct{})
		}
		rm.memberOf[c][roomName] = struct{}{}
	})
}

func (rm *RoomManager) LeaveRoom(roomName string, c *Client) {
	rm.do(func() {
		r := rm.rooms[roomName]
		if r == nil {
			return
		}

		r.Leave(c)

		if m := rm.memberOf[c]; m != nil {
			delete(m, roomName)
			if len(m) == 0 {
				delete(rm.memberOf, c)
			}
		}
	})
}

func (rm *RoomManager) Disconnect(c *Client) {
	rm.do(func() {
		rooms := rm.memberOf[c]
		for roomName := range rooms {
			if r := rm.rooms[roomName]; r != nil {
				r.Leave(c)
			}
		}
		delete(rm.memberOf, c)
	})
}

func (rm *RoomManager) LoadRoomsFromConfig(cfg chatconfig.ChatConfig) {
	for _, rc := range cfg.Rooms {
		opts := []RoomOption{
			WithCapacity(rc.Capacity),
			WithPassword(rc.Password),
			WithPermanent(),
		}
		rm.CreateRoom(rc.Name, opts...)
	}
}
