package session

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	key := generateID()
	data := []byte("tree.xie")
	ttl := 300
	ms, _ := NewMemoryStore(1024)

	t.Run("not init", func(t *testing.T) {
		tmp := &MemoryStore{}
		_, err := tmp.Get(key)
		if err != ErrNotInit {
			t.Fatalf("should return not init error")
		}
		err = tmp.Set(key, data, ttl)
		if err != ErrNotInit {
			t.Fatalf("should return not init error")
		}
		err = tmp.Destroy(key)
		if err != ErrNotInit {
			t.Fatalf("should return not init error")
		}
	})

	t.Run("get not exists data", func(t *testing.T) {
		buf, err := ms.Get(key)
		if err != nil || len(buf) != 0 {
			t.Fatalf("shoud return empty bytes")
		}
	})

	t.Run("set data", func(t *testing.T) {
		err := ms.Set(key, data, ttl)
		if err != nil {
			t.Fatalf("set data fail, %v", err)
		}
		buf, err := ms.Get(key)
		if err != nil {
			t.Fatalf("get data fail after set, %v", err)
		}
		if !bytes.Equal(data, buf) {
			t.Fatalf("the data is not the same after set")
		}
	})

	t.Run("destroy", func(t *testing.T) {
		err := ms.Destroy(key)
		if err != nil {
			t.Fatalf("destory data fail, %v", err)
		}
		buf, err := ms.Get(key)
		if err != nil || len(buf) != 0 {
			t.Fatalf("shoud return empty bytes after destroy")
		}
	})

	t.Run("expired", func(t *testing.T) {
		err := ms.Set(key, data, -100)
		if err != nil {
			t.Fatalf("set data fail, %v", err)
		}
		buf, err := ms.Get(key)
		if err != nil {
			t.Fatalf("get data fail after set, %v", err)
		}
		fmt.Println(buf)
		if len(buf) != 0 {
			t.Fatalf("expired data should be nil")
		}
	})
}
