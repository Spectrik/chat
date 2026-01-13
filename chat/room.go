package chat

import "fmt"

type RoomOption func(*Room)

type Room struct {
	name string

	join      chan *Client
	leave     chan *Client
	broadcast chan string

	clients   map[*Client]struct{}
	capacity  int // 0 = unlimited
	permanent bool
	password  string //Predelat na vic safe vec

	// jednoduchý banlist (key = client.id)
	banList map[string]BanEntry

	// řízení životního cyklu gorutiny roomky
	stop chan struct{}
	done chan struct{}
}

func NewRoom(name string, opts ...RoomOption) *Room {
	r := &Room{
		name:      name,
		join:      make(chan *Client),
		leave:     make(chan *Client),
		broadcast: make(chan string, 128), // buffer, aby sendery tolik neblokovaly

		clients:   make(map[*Client]struct{}),
		banList: make(map[string]BanEntry),

		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

		for _, o := range opts {
		o(r)
	}

	go r.loop()
	return r
}

func (r *Room) loop() {
	defer close(r.done)

	for {
		select {
			case c := <-r.join:
				if r.capacity > 0 && len(r.clients) >= r.capacity {
					c.Out <- fmt.Sprintf("Room %q is full.", r.name)
					continue
				}

				if r.IsBanned(c) {
					c.Out <- fmt.Sprintf("You are banned from room %q.", r.name)
					continue
				}

				r.clients[c] = struct{}{}
				r.Broadcast(fmt.Sprintf("%s has joined the room.", c.Name))

			case c := <-r.leave:
				if _, ok := r.clients[c]; ok {
					delete(r.clients, c)
					r.Broadcast(fmt.Sprintf("%s has left the room.", c.Name))
				}

			case msg := <-r.broadcast:
				for c := range r.clients {
					SendClientMessage(c.Out, msg)
				}

			case <-r.stop:
				r.Broadcast(fmt.Sprintf("Room %s is closing.", r.name))
				return
		}
	}
}

func WithPermanent() RoomOption {
	return func(r *Room) { r.permanent = true }
}

func WithPassword(password string) RoomOption {
	return func(r *Room) { r.password = password }
}

func WithCapacity(capacity int) RoomOption {
	return func(r *Room) { r.capacity = capacity }
}

func SendClientMessage(ch chan string, msg string) {
	select {
		case ch <- msg:
		default:
	}
}

func (r *Room) IsBanned(client *Client) (bool) {
	_, exists := r.banList[client.Conn.RemoteAddr()]
	return exists
}

func (r *Room) Broadcast(msg string) {
	for c := range r.clients {
		SendClientMessage(c.Out, msg)
	}
}

func (r *Room) ClientCount() int {
	return len(r.clients)
}

func (r *Room) Stop() {
	close(r.stop)
	<-r.done
}

func (r *Room) Join(c *Client)  { r.join <- c }

func (r *Room) Leave(c *Client) { r.leave <- c }

func (r *Room) Say(from *Client, text string) {
	r.broadcast <- fmt.Sprintf("[%s] [%s] %s", r.name, from.Name, text)
}
