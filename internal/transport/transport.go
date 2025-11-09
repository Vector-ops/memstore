package transport

type Transport interface {
	Ping() (string, error)
	Send([]byte) (int, error)
	Read() ([]byte, error)
	ReadLoop() error
	GetRemoteAddress() string
	GetLocalAddress() string
}
