package persist

import (
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"
)

type MemoryStore struct {
	Cache *cache.Cache
}

func NewMemoryStore(defaultExpiration time.Duration) *MemoryStore {
	return &MemoryStore{
		Cache: cache.New(defaultExpiration, time.Minute),
	}
}

func (c *MemoryStore) Set(key string, value interface{}, expire time.Duration) error {
	c.Cache.Set(key, value, expire)
	return nil
}

func (c *MemoryStore) Delete(key string) error {
	c.Cache.Delete(key)
	return nil
}

func (c *MemoryStore) Get(key string, value interface{}) error {
	val, found := c.Cache.Get(key)
	if !found {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(val))
	return nil
}
