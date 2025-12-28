package server

import (
	"bytes"
	"log/slog"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/vector-ops/memstore/internal/protocol"
	"github.com/vector-ops/memstore/internal/storage"
	"github.com/vector-ops/memstore/internal/transport"
)

type Server struct {
	id        uuid.UUID
	addr      string
	peers     map[transport.Transport]bool
	ln        net.Listener
	addPeerCh chan transport.Transport
	delPeerCh chan transport.Transport
	quitCh    chan struct{}
	msgCh     chan transport.Message

	kv *storage.KV
	mu *sync.Mutex
	wg *sync.WaitGroup
}

func NewServer(addr string) *Server {
	return &Server{
		id:        uuid.New(),
		addr:      addr,
		peers:     make(map[transport.Transport]bool),
		addPeerCh: make(chan transport.Transport),
		delPeerCh: make(chan transport.Transport),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan transport.Message),
		kv:        storage.NewKeyVal(),
		mu:        &sync.Mutex{},
		wg:        &sync.WaitGroup{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("memstore server running", "listenAddr", s.addr)

	return s.acceptLoop()
}

func (s *Server) handleMsg(msg transport.Message) error {
	switch v := msg.Cmd.(type) {
	case protocol.SetCommand:
		if err := s.kv.Set(v.Key, v.Value); err != nil {
			return err
		}

	case protocol.GetCommand:
		var err error
		val, ok := s.kv.Get(v.Key)
		if !ok {
			val = []byte("nil")
		}
		val = append(val, '\n')
		_, err = msg.Transport.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

	case protocol.KeysCommand:
		var buf bytes.Buffer
		keys, ok := s.kv.Keys()
		if ok {
			if _, err := buf.Write(keys); err != nil {
				slog.Error("keys not found on master", "err", err)
			}
		}

		_, err := msg.Transport.Send(buf.Bytes())
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

	case protocol.PingCommand:
		var buf bytes.Buffer
		if _, err := buf.WriteString(s.id.String()); err != nil {
			slog.Error("Failed to write to buffer", "err", err)
		}
		if err := buf.WriteByte('\n'); err != nil {
			slog.Error("Failed to write to buffer", "err", err)
		}

		_, err := msg.Transport.Send(buf.Bytes())
		if err != nil {
			slog.Error("peer send eror", "err", err)
		}
	}
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case msg := <-s.msgCh:
			if err := s.handleMsg(msg); err != nil {
				slog.Error("raw message error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			slog.Info("new peer connected", "remoteAddr", peer.GetRemoteAddress())
			s.peers[peer] = true

		case peer := <-s.delPeerCh:
			slog.Info("peer disconnected", "remoteAddr", peer.GetRemoteAddress())
			delete(s.peers, peer)
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error: ", "err", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := transport.NewTCPTransport(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer
	if err := peer.ReadLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}
