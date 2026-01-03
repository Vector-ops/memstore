// Package server provides the main server that handles the KV store and participates in cluster operations
package server

import (
	"bytes"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vector-ops/memstore/internal/cluster"
	"github.com/vector-ops/memstore/internal/protocol"
	"github.com/vector-ops/memstore/internal/storage"
	"github.com/vector-ops/memstore/internal/transport"
)

type ServerConfig struct {
	ServerAddr string
	// ClusterAddr string
	LeaderAddr string

	// leader config
	HeartbeatInterval     time.Duration
	ReplicationTimeout    time.Duration
	ReplicationMaxRetries int
	ReplicaCount          int

	// follower config
	HeartbeatTimeout time.Duration
}

type Server struct {
	// server internals
	id       uuid.UUID
	cfg      ServerConfig
	hbTicker *time.Ticker
	mu       *sync.Mutex
	wg       *sync.WaitGroup

	// network
	peers     map[transport.Transport]bool
	ln        net.Listener
	addPeerCh chan transport.Transport
	delPeerCh chan transport.Transport
	quitCh    chan struct{}
	msgCh     chan transport.Message

	// cluster
	role     cluster.Role
	clstrMgr *cluster.ClusterManager

	// database
	kv *storage.KV
}

func NewServer(cfg ServerConfig) *Server {

	clstrMgr := cluster.NewClusterManager(cluster.FOLLOWER, cfg.ReplicaCount)

	// validate config
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = time.Millisecond * 50
	}

	if cfg.HeartbeatTimeout == 0 {
		cfg.HeartbeatTimeout = time.Millisecond * time.Duration(rand.Int63n(300))
	}

	if cfg.ReplicationMaxRetries == 0 {
		cfg.ReplicationMaxRetries = 2
	}

	if cfg.ReplicationTimeout == 0 {
		cfg.ReplicationTimeout = time.Second * 5
	}

	if cfg.ServerAddr == "" {
		log.Fatalf("server address not found in config")
	}

	return &Server{
		id:        uuid.New(),
		cfg:       cfg,
		peers:     make(map[transport.Transport]bool),
		addPeerCh: make(chan transport.Transport),
		delPeerCh: make(chan transport.Transport),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan transport.Message),
		kv:        storage.NewKeyVal(),
		mu:        &sync.Mutex{},
		wg:        &sync.WaitGroup{},

		clstrMgr: clstrMgr,
		role:     cluster.FOLLOWER,
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.cfg.ServerAddr)
	if err != nil {
		return err
	}
	s.ln = ln

	if s.role == cluster.FOLLOWER {
		s.hbTicker = time.NewTicker(s.cfg.HeartbeatTimeout)
	}

	go s.loop()

	slog.Info("memstore server running", "listenAddr", s.cfg.ServerAddr)

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
		s.hbTicker.Reset(s.cfg.HeartbeatTimeout)
		log.Println("Recieved heartbeat resetting timer")
		// msg.Transport.Close()
		// log.Println("Closed safely")
		// var buf bytes.Buffer
		// if _, err := buf.WriteString(s.id.String()); err != nil {
		// 	slog.Error("Failed to write to buffer", "err", err)
		// }
		// if err := buf.WriteByte('\n'); err != nil {
		// 	slog.Error("Failed to write to buffer", "err", err)
		// }

		// _, err := msg.Transport.Send(buf.Bytes())
		// if err != nil {
		// 	slog.Error("peer send eror", "err", err)
		// }
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
		case t := <-s.hbTicker.C:
			// TODO: send vote request with the new term to all the nodes
			if s.role == cluster.FOLLOWER {
				s.role = cluster.CANDIDATE
				log.Println("Heartbeat timedout: turning into candidate", t)
				s.hbTicker.Reset(s.cfg.HeartbeatTimeout)
			}
		case <-s.quitCh:
			s.Stop()
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

func (s *Server) StartHearbeat() {
	// TODO: send hearbeat ping to all nodes in the cluster
}

func (s *Server) stopHeartbeat() {}

func (s *Server) Stop() {
	s.hbTicker.Stop()
	s.ln.Close()

	close(s.addPeerCh)
	close(s.delPeerCh)
	close(s.msgCh)
	close(s.quitCh)

}
