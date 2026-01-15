package client

import (
	"sync"

	"github.com/ondrej/chat/internal/user"
	"github.com/ondrej/chat/transport"
)

type Identity struct {
	User    *user.User    // Registered
	AnonymousName string
	Ip		string
}
type Client struct {
	Conn transport.Transport
	Out  chan string
	Quit bool
	done chan struct{}
	once sync.Once
	Identity Identity
}

func NewClient(conn transport.Transport, name string) *Client {
	return &Client{
		Conn: conn,
		Out:  make(chan string, 32),
		done: make(chan struct{}),
		Identity: Identity{AnonymousName: name, Ip: conn.RemoteAddr()},
	}
}

func (c *Client) Close() {
	c.once.Do(func() { close(c.done) })
}

func (c *Client) ClientName() string{
	if c.Authenticated() {return c.Identity.User.Username}
	return c.Identity.AnonymousName
}

func (c *Client) Authenticated() bool {
	return c.Identity.User != nil
}

func (c *Client) Done() <-chan struct{} { return c.done }

func (c *Client) Send(msg string) bool {
	select {
		case <-c.done:
			return false
		case c.Out <- msg:
			return true
		default:
			return false
	}
}
