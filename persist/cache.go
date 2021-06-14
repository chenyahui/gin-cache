package persist

import (
	"errors"
	"time"
)

var ErrCacheMiss = errors.New("persist cache miss error")

// CacheStore is the interface of a Cache backend
type CacheStore interface {
	// Get retrieves an item from the Cache. Returns the item or nil, and a bool indicating
	// whether the key was found.
	Get(key string, value interface{}) error

	// Set sets an item to the Cache, replacing any existing item.
	Set(key string, value interface{}, expire time.Duration) error

	// Delete removes an item from the Cache. Does nothing if the key is not in the Cache.
	Delete(key string) error
}
