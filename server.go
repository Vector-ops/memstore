package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Config struct {
	NodeConfig

	Replicas []NodeConfig
}

type Message struct {
	cmd  Command
	peer *Peer
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	delPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message

	kv       *KV
	hashRing *HashRing

	mu *sync.Mutex
	wg *sync.WaitGroup
}

func NewServer(cfg Config) *Server {
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		hashRing:  NewHashRing(),
		addPeerCh: make(chan *Peer),
		delPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKeyVal(),
		mu:        &sync.Mutex{},
		wg:        &sync.WaitGroup{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.Host, s.Port))
	if err != nil {
		return err
	}
	s.ln = ln

	go s.loop()

	slog.Info("memstore server running", "listenAddr", s.Host+":"+s.Port)

	s.registerReplicas(context.Background())

	return s.acceptLoop()
}

func (s *Server) handleMsg(msg Message) error {
	switch v := msg.cmd.(type) {
	case SetCommand:
		if err := s.kv.Set(v.key, v.value); err != nil {
			return err
		}
		if err := s.hashRing.StoreKey(string(v.key), v.value); err != nil {
			slog.Error("node store error", "err", err)
		}

	case GetCommand:
		var err error
		val, ok := s.kv.Get(v.key)
		if !ok {
			val, err = s.hashRing.RetrieveKey(string(v.key))
			if err != nil {
				val = []byte("nil")
			}
		}
		val = append(val, '\n')
		_, err = msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}

	case KeysCommand:
		var buf bytes.Buffer
		keys, ok := s.kv.Keys()
		if ok {
			if _, err := buf.Write(keys); err != nil {
				slog.Error("keys not found on master", "err", err)
			}
		}

		nodeKeys, err := s.hashRing.RetrieveKeys()
		if err == nil {
			if _, err := buf.Write(nodeKeys); err != nil {
				slog.Error("keys not found on master", "err", err)
			}
		}

		_, err = msg.peer.Send(buf.Bytes())
		if err != nil {
			slog.Error("peer send error", "err", err)
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
			slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())
			s.peers[peer] = true

		case peer := <-s.delPeerCh:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())
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
	peer := NewPeer(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}

func (s *Server) registerReplicas(ctx context.Context) {
	for _, config := range s.Config.Replicas {
		go func(config NodeConfig) {
			for {
				addr := formatHostPort(config.Host, config.Port)
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					// log.Printf("Failed to connect to replica %s: %v\n", config.Id, err)
					// log.Println("Trying again in 5s")
					time.Sleep(time.Second * 5)
					continue
				}
				peer := NewPeer(conn, s.msgCh, s.delPeerCh)
				s.hashRing.AddNode(config.Id, peer)
				log.Printf("Connected to replica %s", conn.RemoteAddr())
				break
			}
		}(config)
	}
}
