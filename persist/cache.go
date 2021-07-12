package persist

import (
	"errors"
	"time"
)

// ErrCacheMiss represent the cache key does not exist in the store
var ErrCacheMiss = errors.New("persist cache miss error")

// CacheStore is the interface of a Cache backend
type CacheStore interface {
	// Get retrieves an item from the Cache. if key does not exist in the store, return ErrCacheMiss
	Get(key string, value interface{}) error

	// Set sets an item to the Cache, replacing any existing item.
	Set(key string, value interface{}, expire time.Duration) error

	// Delete removes an item from the Cache. Does nothing if the key is not in the Cache.
	Delete(key string) error
}
