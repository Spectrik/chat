package transport

type Transport interface {
	Read() (string, error)
	Write(msg string) error
	Close() error
	RemoteAddr() string
}
