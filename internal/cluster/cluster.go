package cluster

import (
	"errors"

	"github.com/vector-ops/memstore/internal/config"
	"github.com/vector-ops/memstore/internal/server"
	"github.com/vector-ops/memstore/internal/transport"
)

var ErrTooManyNodes = errors.New("Cluster Error: Replica count exceeds configured count")

type ClusterManager struct {
	config.ServerConfig

	currentNode *Node

	// replication
	replicas map[string]*Node
	hr       *HashRing
}

func NewClusterManager(cfg config.ServerConfig) *ClusterManager {
	var role Role
	if cfg.LeaderAddr == cfg.ClusterAddr {
		role = LEADER
	}

	currentNode := NewNode(cfg.Id, cfg.ServerAddr, role, nil)

	s := server.NewServer(cfg.ServerAddr)

	currentNode.SetServer(s)

	return &ClusterManager{
		ServerConfig: cfg,

		currentNode: currentNode,
		replicas:    make(map[string]*Node),
		hr:          NewHashRing(),
	}
}

func (cm *ClusterManager) Start() error {
	return cm.currentNode.server.Start()
}

func (cm *ClusterManager) AddNode(connectionString, id string, role Role) error {
	if cm.ReplicaCount > len(cm.replicas) {
		return ErrTooManyNodes
	}

	t := transport.NewHttpTransport(connectionString)

	node := &Node{
		ConnectionString: connectionString,
		ID:               id,
		Transport:        t,
		Role:             role,
	}

	if err := cm.hr.AddNode(t); err != nil {
		return err
	}

	cm.replicas[id] = node

	go node.startHeartBeat()

	return nil
}

func (cm *ClusterManager) RemoveNode(id string) error {
	cm.replicas[id].stopHeartBeat()
	cm.hr.RemoveNode(id)
	delete(cm.replicas, id)

	return nil
}
