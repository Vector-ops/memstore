package transport

import "github.com/vector-ops/memstore/internal/protocol"

type Message struct {
	Cmd       protocol.Command
	Transport Transport
}
