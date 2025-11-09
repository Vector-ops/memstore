package storage

import (
	"strings"
	"sync"
)

type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewKeyVal() *KV {
	return &KV{
		data: make(map[string][]byte),
	}
}

func (kv *KV) Set(key, value []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[string(key)] = value
	return nil
}

func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	val, ok := kv.data[string(key)]
	return val, ok
}

func (kv *KV) Keys() ([]byte, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	keys := []string{}
	for k := range kv.data {
		keys = append(keys, k)
	}

	keysString := strings.Join(keys, ",")

	return []byte(keysString), true
}
