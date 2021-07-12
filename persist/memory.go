package persist

import (
	"errors"
	"github.com/ReneKroon/ttlcache/v2"
	"reflect"
	"time"
)

// MemoryStore local memory cache store
type MemoryStore struct {
	Cache *ttlcache.Cache
}

// NewMemoryStore allocate a local memory store with default expiration
func NewMemoryStore(defaultExpiration time.Duration) *MemoryStore {
	cacheStore := ttlcache.NewCache()
	_ = cacheStore.SetTTL(defaultExpiration)
	return &MemoryStore{
		Cache: cacheStore,
	}
}

func (c *MemoryStore) Set(key string, value interface{}, expire time.Duration) error {
	return c.Cache.SetWithTTL(key, value, expire)
}

func (c *MemoryStore) Delete(key string) error {
	return c.Cache.Remove(key)
}

func (c *MemoryStore) Get(key string, value interface{}) error {
	val, err := c.Cache.Get(key)
	if errors.Is(err, ttlcache.ErrNotFound) {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(val))
	return nil
}
