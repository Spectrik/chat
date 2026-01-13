package chat

import (
	transport "github.com/ondrej/chat/transport"
)

type Client struct {
	Conn transport.Transport
	Name string
	Out  chan string
	Quit bool
}

func NewClient(conn transport.Transport, name string) *Client {
	return &Client{
		Conn: conn,
		Name: name,
		Out:  make(chan string, 32),
	}
}

func HelpMessage(out chan <- string) {
	out <- "Available commands:"
	out <- "/help - Show this help message"
	out <- "/join <roomname> - Join a room"
	out <- "/leave <roomname> - Leave a room"
	out <- "/roomlist - List rooms you are in"
	out <- "/say <roomname> <message> - Send message to a room"
	out <- "/quit - Quit the chat"
}
