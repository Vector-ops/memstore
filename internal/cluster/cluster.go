// Package cluster provides an interface to communnicate and manage other nodes on the cluster
package cluster

import (
	"errors"

	"github.com/vector-ops/memstore/internal/transport"
)

var ErrTooManyNodes = errors.New("cluster error: replica count exceeds configured count")

type ClusterManager struct {
	hbChan chan bool

	// replication
	replicaCount int
	replicas     map[string]*Node
	hr           *HashRing
}

func NewClusterManager(role Role, replicaCount int) *ClusterManager {

	return &ClusterManager{
		replicaCount: replicaCount,
		replicas:     make(map[string]*Node),
		hr:           NewHashRing(),
	}
}

func (cm *ClusterManager) AddNode(connectionString, id string, role Role) error {
	if cm.replicaCount > len(cm.replicas) {
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

func (cm *ClusterManager) StartHeartbeat() {
	cm.hbChan = make(chan bool)
	go func() {
		select {
		case <-cm.hbChan:
			return
		default:
		}

		// TODO: send heartbeat pings
	}()
}

func (cm *ClusterManager) StopHeartbeat() {
	close(cm.hbChan)
}
