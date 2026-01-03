package transport

type HttpTransport struct {
	connectionStr string
}

// Close implements [Transport].
func (h *HttpTransport) Close() {
	panic("unimplemented")
}

func NewHttpTransport(connectionStr string) Transport {
	return &HttpTransport{
		connectionStr: connectionStr,
	}
}

// GetLocalAddress implements [Transport].
func (h *HttpTransport) GetLocalAddress() string {
	return h.connectionStr
}

// GetRemoteAddress implements [Transport].
func (h *HttpTransport) GetRemoteAddress() string {
	return h.connectionStr
}

// Ping implements [Transport].
func (h *HttpTransport) Ping() (string, error) {
	panic("unimplemented")
}

// Read implements [Transport].
func (h *HttpTransport) Read() ([]byte, error) {
	panic("unimplemented")
}

// ReadLoop implements [Transport].
func (h *HttpTransport) ReadLoop() error {
	panic("unimplemented")
}

// Send implements [Transport].
func (h *HttpTransport) Send([]byte) (int, error) {
	panic("unimplemented")
}
