package kv

import (
	"sync"
)

func NewMemoryStore() Store {
	return &MemoryStore{}
}

// MemoryStore is an in-memory KV store
type MemoryStore struct {
	store sync.Map
}

// Get implements Get for MemoryStore from the Store interface
func (ms *MemoryStore) Get(key []byte) (value []byte, err error) {
	v, ok := ms.store.Load(string(key))
	if !ok {
		return nil, ErrNoExist
	}
	return v.([]byte), nil
}

// Set implements Set for MemoryStore from the Store interface
func (ms *MemoryStore) Set(key []byte, value []byte) (err error) {
	ms.store.Store(string(key), value)
	return nil
}
