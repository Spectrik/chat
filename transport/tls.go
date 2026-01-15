package transport

import (
	"bufio"
	"crypto/tls"
	"io"
)

type TLSTransport struct {
	conn *tls.Conn
	w *bufio.Writer
	sc *bufio.Scanner
}

func NewTLSTransport(c *tls.Conn) *TLSTransport {
    sc := bufio.NewScanner(c)
    sc.Buffer(make([]byte, 1024), 1024*1024)

    return &TLSTransport{
        conn: c,
        sc: sc,
        w:  bufio.NewWriter(c),
    }
}

func (t *TLSTransport) Read() (string, error) {
    if !t.sc.Scan() {
        if err := t.sc.Err(); err != nil {
            return "", err
        }
        return "", io.EOF
    }

    return t.sc.Text(), nil
}

func (t *TLSTransport) Write(msg string) error {
    _, err := t.w.WriteString(msg + "\n")
    if err != nil {
        return err
    }

    return t.w.Flush()
}

func (t *TLSTransport) Close() error { return t.conn.Close() }

func (t *TLSTransport) RemoteAddr() string {
    return t.conn.RemoteAddr().String()
}
