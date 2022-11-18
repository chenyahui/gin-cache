package persist

import (
	"errors"
	"reflect"
	"time"

	"github.com/jellydator/ttlcache/v2"
)

// MemoryStore local memory cache store
type MemoryStore struct {
	cache *ttlcache.Cache
}

// NewMemoryStore allocate a local memory store with default expiration
func NewMemoryStore(defaultExpiration time.Duration) *MemoryStore {
	cacheStore := ttlcache.NewCache()
	_ = cacheStore.SetTTL(defaultExpiration)

	// disable SkipTTLExtensionOnHit by default
	cacheStore.SkipTTLExtensionOnHit(true)

	return &MemoryStore{
		cache: cacheStore,
	}
}

// Set put key value pair to memory store, and expire after expireDuration
func (c *MemoryStore) Set(key string, value interface{}, expireDuration time.Duration) error {
	return c.cache.SetWithTTL(key, value, expireDuration)
}

// Delete remove key in memory store, do nothing if key doesn't exist
func (c *MemoryStore) Delete(key string) error {
	return c.cache.Remove(key)
}

// Get key in memory store, if key doesn't exist, return ErrCacheMiss
func (c *MemoryStore) Get(key string, value interface{}) error {
	val, err := c.cache.Get(key)
	if errors.Is(err, ttlcache.ErrNotFound) {
		return ErrCacheMiss
	}

	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(val))
	return nil
}

// SetCacheSizeLimit sets a limit to the amount of cached items.
// If a new item is getting cached, the closes item to being timed out will be replaced
// Set to 0 to turn off
func (c *MemoryStore) SetCacheSizeLimit(limit int) {
	c.cache.SetCacheSizeLimit(limit)
}
