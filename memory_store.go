package session

import (
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

var (
	// ErrNotInit error not init
	ErrNotInit = errors.New("client not init")
)

type (
	// MemoryStore memory store for session
	MemoryStore struct {
		client *lru.Cache
	}
	// MemoryStoreInfo memory store info
	MemoryStoreInfo struct {
		ExpiredAt int64
		Data      []byte
	}
)

// Get get the seesion from memory
func (ms *MemoryStore) Get(key string) (data []byte, err error) {
	client := ms.client
	if client == nil {
		err = ErrNotInit
		return
	}
	v, found := client.Get(key)
	if !found {
		return
	}
	info, ok := v.(*MemoryStoreInfo)
	if !ok {
		return
	}
	if info.ExpiredAt < time.Now().Unix() {
		return
	}
	data = info.Data
	return
}

// Set set the sesion to memory
func (ms *MemoryStore) Set(key string, data []byte, ttl int) (err error) {
	client := ms.client
	if client == nil {
		err = ErrNotInit
		return
	}
	expiredAt := time.Now().Unix() + int64(ttl)
	info := &MemoryStoreInfo{
		ExpiredAt: expiredAt,
		Data:      data,
	}
	client.Add(key, info)
	return
}

// Destroy remove the session from memory
func (ms *MemoryStore) Destroy(key string) (err error) {
	client := ms.client
	if client == nil {
		err = ErrNotInit
		return
	}
	client.Remove(key)
	return
}

// NewMemoryStore create new memory store instance
func NewMemoryStore(size int) (store *MemoryStore, err error) {
	client, err := lru.New(size)
	if err != nil {
		return
	}
	store = &MemoryStore{
		client: client,
	}
	return
}
