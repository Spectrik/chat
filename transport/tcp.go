package transport

import (
	"bufio"
	"io"
	"net"
)

type TCPTransport struct {
	conn net.Conn
	w *bufio.Writer
	sc *bufio.Scanner
}

func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{conn: conn, w: bufio.NewWriter(conn), sc: bufio.NewScanner(conn)}
}

func (t *TCPTransport) Read() (string, error) {
	if !t.sc.Scan() {
		if err := t.sc.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}

	return t.sc.Text(), nil
}


func (t *TCPTransport) Write(msg string) error {
	_, err := t.w.WriteString(msg + "\n")
	if err != nil {
		return err
	}

	return t.w.Flush()
}

func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

func (t *TCPTransport) RemoteAddr() string {
	return t.conn.RemoteAddr().String()
}

func (t *TCPTransport) Scanner() *bufio.Scanner {
	return t.sc
}
