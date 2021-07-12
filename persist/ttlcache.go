package persist

import (
	"errors"
	"github.com/ReneKroon/ttlcache/v2"
	"reflect"
	"time"
)

type TTLMemoryStore struct {
	Cache *ttlcache.Cache
}

func NewTTLMemoryStore(defaultExpiration time.Duration) *TTLMemoryStore {
	cacheStore := ttlcache.NewCache()
	_ = cacheStore.SetTTL(defaultExpiration)
	return &TTLMemoryStore{
		Cache: cacheStore,
	}
}

func (c *TTLMemoryStore) Set(key string, value interface{}, expire time.Duration) error {
	return c.Cache.SetWithTTL(key, value, expire)
}

func (c *TTLMemoryStore) Delete(key string) error {
	return c.Cache.Remove(key)
}

func (c *TTLMemoryStore) Get(key string, value interface{}) error {
	val, err := c.Cache.Get(key)
	if errors.Is(err, ttlcache.ErrNotFound) {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(val))
	return nil
}
