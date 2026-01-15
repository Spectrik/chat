package transport

import "golang.org/x/net/websocket"

type WSTransport struct {
	conn *websocket.Conn
}

// TODO