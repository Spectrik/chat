package transport

import (
	"golang.org/x/net/websocket"
)

type Transport interface {
	Read() (string, error)
	Write(msg string) error
	Close() error
	RemoteAddr() string
}

type WSTransport struct {
	conn *websocket.Conn
}
