package session

import (
	"bytes"
	"testing"

	"github.com/go-redis/redis"
)

func TestRedisStore(t *testing.T) {
	key := generateID("")
	data := []byte("tree.xie")
	ttl := 300
	rs := NewRedisStore(nil, &redis.Options{
		Addr: "localhost:6379",
	})
	t.Run("get not exists data", func(t *testing.T) {
		buf, err := rs.Get(key)
		if err != nil || len(buf) != 0 {
			t.Fatalf("shoud return empty bytes")
		}
	})

	t.Run("set data", func(t *testing.T) {
		err := rs.Set(key, data, ttl)
		if err != nil {
			t.Fatalf("set data fail, %v", err)
		}
		buf, err := rs.Get(key)
		if err != nil {
			t.Fatalf("get data fail after set, %v", err)
		}
		if !bytes.Equal(data, buf) {
			t.Fatalf("the data is not the same after set")
		}
	})

	t.Run("destroy", func(t *testing.T) {
		err := rs.Destroy(key)
		if err != nil {
			t.Fatalf("destory data fail, %v", err)
		}
		buf, err := rs.Get(key)
		if err != nil || len(buf) != 0 {
			t.Fatalf("shoud return empty bytes after destroy")
		}
	})

	// err := rs.Set(key, data, ttl)

}
