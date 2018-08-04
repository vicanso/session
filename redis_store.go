package session

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

type (
	// RedisStore redis store for session
	RedisStore struct {
		client *redis.Client
	}
)

// Get get the session from redis
func (rs *RedisStore) Get(key string) ([]byte, error) {
	buf, err := rs.client.Get(key).Bytes()
	if err == redis.Nil {
		return buf, nil
	}
	return buf, err
}

// Set set the session to redis
func (rs *RedisStore) Set(key string, data []byte, ttl int) error {
	expiration := time.Duration(int64(time.Second) * int64(ttl))
	return rs.client.Set(key, data, expiration).Err()
}

// Destroy remove the session from redis
func (rs *RedisStore) Destroy(key string) error {
	return rs.client.Del(key).Err()
}

// NewRedisStore create new redis store instance
func NewRedisStore(client *redis.Client, opts *redis.Options) *RedisStore {
	if client == nil && opts == nil {
		panic(errors.New("client and opts can both be nil"))
	}
	rs := &RedisStore{}
	if client != nil {
		rs.client = client
	} else {
		rs.client = redis.NewClient(opts)
	}
	return rs
}
