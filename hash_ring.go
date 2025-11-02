package main

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"slices"
	"sync"

	"github.com/tidwall/resp"
)

var (
	ErrNodeNotFound   = errors.New("Node(s) not found")
	ErrStoreFailed    = errors.New("Failed to store key on node")
	ErrRetreiveFailed = errors.New("Failed to retrieve key from node")
)

type Node struct {
	ID   string
	Peer *Peer
}

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

func (r *HashRing) AddNode(id string, p *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hash := r.hash(id)
	node := &Node{
		ID:   id,
		Peer: p,
	}
	r.nodes[hash] = node
	r.hashes = append(r.hashes, hash)
	slices.Sort(r.hashes)
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
	// copy(node.peer.keys, nextNode.peer.keys)
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
		resp.StringValue(CommandSET),
		resp.StringValue(key),
		resp.StringValue(string(value)),
	})

	_, err := node.Peer.Send(buf.Bytes())

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
		resp.StringValue(CommandGET),
		resp.StringValue(key),
	})

	_, err := node.Peer.Send(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("node id: %s, %w", node.ID, err)
	}

	value, err := node.Peer.Read()
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
		resp.StringValue(CommandKEYS),
	})

	var resBuf bytes.Buffer
	for _, node := range r.nodes {
		_, err := node.Peer.Send(buf.Bytes())
		if err != nil {
			slog.Error(fmt.Sprintf("node id: %s", node.ID), "err", err)
			continue
		}

		value, err := node.Peer.Read()
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
