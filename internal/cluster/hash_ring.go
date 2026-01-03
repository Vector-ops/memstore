package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"slices"
	"sync"

	"github.com/tidwall/resp"
	"github.com/vector-ops/memstore/internal/protocol"
	"github.com/vector-ops/memstore/internal/transport"
)

var (
	ErrNodeNotFound   = errors.New("node(s) not found")
	ErrStoreFailed    = errors.New("failed to store key on node")
	ErrRetreiveFailed = errors.New("failed to retrieve key from node")
)

type HashRing struct {
	mu     *sync.RWMutex
	hashes []uint32
	nodes  map[uint32]*Node
}

func NewHashRing() *HashRing {
	return &HashRing{
		mu:     &sync.RWMutex{},
		hashes: make([]uint32, 0),
		nodes:  make(map[uint32]*Node),
	}
}

func (r *HashRing) hash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

func (r *HashRing) AddNode(p transport.Transport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peerID, err := p.Ping()
	if err != nil {
		slog.Error("Failed to add node", "node remote address", p.GetRemoteAddress(), "error", err)
		return err
	}

	slog.Info("Successful ping", "node id: ", peerID)

	hash := r.hash(peerID)
	node := &Node{
		ID:        peerID,
		Transport: p,
	}
	r.nodes[hash] = node
	r.hashes = append(r.hashes, hash)
	slices.Sort(r.hashes)

	return nil
}

func (r *HashRing) nextNodeIndex(hash uint32) int {
	for i, h := range r.hashes {
		if h > hash {
			return i
		}
	}

	return 0
}

func (r *HashRing) RemoveNode(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hash := r.hash(id)
	node := r.nodes[hash]
	nextNodeIndex := r.nextNodeIndex(hash)
	nextNode := r.nodes[r.hashes[nextNodeIndex]]

	// bulk copy all current node's keys to the next node
	// copy(node.Transport.keys, nextNode.Transport.keys)
	_, _ = node, nextNode

	// remove node from map
	delete(r.nodes, hash)

	// remove hash from ring
	for i, h := range r.hashes {
		if h == hash {
			r.hashes = slices.Delete(r.hashes, i, i+1)
			break
		}
	}

	// rebalance ring
	slices.Sort(r.hashes)
}

func (r *HashRing) GetNode(key string) *Node {
	if len(r.hashes) == 0 {
		return nil
	}

	hash := r.hash(key)

	i, _ := slices.BinarySearch(r.hashes, hash)

	if i == len(r.hashes) {
		i = 0
	}

	return r.nodes[r.hashes[i]]
}

func (r *HashRing) StoreKey(key string, value []byte) error {
	node := r.GetNode(key)
	if node == nil {
		return ErrNodeNotFound
	}

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue(protocol.CommandSET),
		resp.StringValue(key),
		resp.StringValue(string(value)),
	})

	_, err := node.Transport.Send(buf.Bytes())

	return err
}

func (r *HashRing) RetrieveKey(key string) ([]byte, error) {
	node := r.GetNode(key)
	if node == nil {
		return nil, ErrNodeNotFound
	}

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue(protocol.CommandGET),
		resp.StringValue(key),
	})

	_, err := node.Transport.Send(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("node id: %s, %w", node.ID, err)
	}

	value, err := node.Transport.Read()
	if err != nil {
		return nil, fmt.Errorf("node id: %s, %w", node.ID, err)
	}

	return value, nil
}

func (r *HashRing) RetrieveKeys() ([]byte, error) {
	if len(r.hashes) == 0 {
		return nil, ErrNodeNotFound
	}

	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	wr.WriteArray([]resp.Value{
		resp.StringValue(protocol.CommandKEYS),
	})

	var resBuf bytes.Buffer
	for _, node := range r.nodes {
		_, err := node.Transport.Send(buf.Bytes())
		if err != nil {
			slog.Error(fmt.Sprintf("node id: %s", node.ID), "err", err)
			continue
		}

		value, err := node.Transport.Read()
		if err != nil {
			slog.Error(fmt.Sprintf("node id: %s", node.ID), "err", err)
			continue
		}

		if _, err = resBuf.Write(value); err != nil {
			slog.Error(fmt.Sprintf("node id: %s", node.ID), "err", err)
		}
	}

	return resBuf.Bytes(), nil
}
