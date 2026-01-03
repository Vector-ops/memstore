package cluster

import (
	"github.com/vector-ops/memstore/internal/transport"
)

type Role string

const (
	FOLLOWER  Role = "follower"
	CANDIDATE Role = "candidate"
	LEADER    Role = "leader"
)

type Node struct {
	ID               string
	ConnectionString string
	Transport        transport.Transport // replica transport
	Role             Role
	stopChan         chan bool
}

func NewNode(id, connectionString string, role Role, t transport.Transport) *Node {
	return &Node{
		ID:               id,
		ConnectionString: connectionString,
		Role:             role,
		Transport:        t,
	}
}

func (n *Node) SetRole(r Role) {
	n.Role = r
}

func (n *Node) startHeartBeat() {
	n.stopChan = make(chan bool)
	for {
		select {
		case <-n.stopChan:
			return
		default:
			n.Transport.Ping()
		}
	}
}

func (n *Node) stopHeartBeat() {
	close(n.stopChan)
}
