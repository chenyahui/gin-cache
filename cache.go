package cache

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

type Options struct {
	// CacheStore the cache backend to store response
	CacheStore persist.CacheStore

	// CacheDuration
	CacheDuration time.Duration

	// DisableSingleFlight means whether use singleflight to avoid Hotspot Invalid when cache miss
	DisableSingleFlight bool

	// SingleflightTimeout this option only be effective when DisableSingleFlight is false
	SingleflightTimeout time.Duration

	// Logger
	Logger Logger
}

type Logger interface {
	Printf(format string, args ...interface{})
}

type Handler func(c *gin.Context) (string, bool)

// Cache user must pass getCacheKey to describe the way to generate cache key
func Cache(handler Handler, options Options) gin.HandlerFunc {
	if options.CacheStore == nil {
		panic("CacheStore can not be nil")
	}

	cacheManager := newCacheManager()

	return func(c *gin.Context) {
		cacheKey, needCache := handler(c)
		if !needCache {
			c.Next()
			return
		}

		respCache := cacheManager.getResponseCache()
		respCache.reset()

		err := options.CacheStore.Get(cacheKey, &respCache)

		defer cacheManager.putResponseCache(respCache)

		if err == nil {
			c.Writer.WriteHeader(respCache.Status)
			for k, vals := range respCache.Header {
				for _, v := range vals {
					c.Writer.Header().Set(k, v)
				}
			}

			if _, err := c.Writer.Write(respCache.Data); err != nil {
				if options.Logger != nil {
					options.Logger.Printf("write response error: %v", err)
				}
			}

			// abort handler chain and return directly
			c.Abort()
			return
		}

		if err != persist.ErrCacheMiss {
			if options.Logger != nil {
				options.Logger.Printf("get cache: %v", err)
			}
		}

		if options.DisableSingleFlight {
			cacheManager.responseWithCache(c, cacheKey, options)
		} else {
			// use singleflight to avoid Hotspot Invalid
			cacheManager.sfGroup.Do(cacheKey, func() (interface{}, error) {
				if options.SingleflightTimeout > 0 {
					go func() {
						time.Sleep(options.SingleflightTimeout)
						cacheManager.sfGroup.Forget(cacheKey)
					}()
				}

				if err := cacheManager.responseWithCache(c, cacheKey, options); err != nil {
					return nil, err
				}

				return nil, nil
			})
		}

	}
}

// CacheByURI a shortcut function for caching response with uri
func CacheByURI(options Options) gin.HandlerFunc {
	return Cache(
		func(c *gin.Context) (string, bool) {
			return c.Request.RequestURI, true
		},
		options,
	)
}

// CacheByPath a shortcut function for caching response with url path, discard the query params
func CacheByPath(options Options) gin.HandlerFunc {
	return Cache(
		func(c *gin.Context) (string, bool) {
			return c.Request.URL.Path, true
		},
		options,
	)
}

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func (c *responseCache) reset() {
	c.Data = c.Data[0:0]
	c.Header = make(http.Header)
}

func (c *responseCache) fill(cacheWriter *cacheWriter) {
	c.Status = cacheWriter.Status()
	c.Data = cacheWriter.body.Bytes()
	c.Header = make(http.Header, len(cacheWriter.Header()))

	for key, value := range cacheWriter.Header() {
		c.Header[key] = value
	}
}

// cacheWriter
type cacheWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *cacheWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *cacheWriter) reset(writer gin.ResponseWriter) {
	w.body.Reset()
	w.ResponseWriter = writer
}

func newCacheWriterPool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return &cacheWriter{}
		},
	}
}

func newResponseCachePool() *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return &responseCache{
				Header: make(http.Header),
			}
		},
	}
}

type cacheManager struct {
	sfGroup           singleflight.Group
	responseCachePool *sync.Pool
	cacheWriterPool   *sync.Pool
}

func newCacheManager() *cacheManager {
	return &cacheManager{
		sfGroup:           singleflight.Group{},
		responseCachePool: newResponseCachePool(),
		cacheWriterPool:   newCacheWriterPool(),
	}
}

func (m *cacheManager) getResponseCache() *responseCache {
	return m.responseCachePool.Get().(*responseCache)
}

func (m *cacheManager) putResponseCache(c *responseCache) {
	m.responseCachePool.Put(c)
}

func (m *cacheManager) getCacheWriter() *cacheWriter {
	return m.cacheWriterPool.Get().(*cacheWriter)
}

func (m *cacheManager) putCacheWriter(w *cacheWriter) {
	m.responseCachePool.Put(w)
}

func (m *cacheManager) responseWithCache(
	c *gin.Context,
	cacheKey string,
	options Options,
) error {
	cacheWriter := m.cacheWriterPool.Get().(*cacheWriter)
	cacheWriter.reset(c.Writer)

	// give back object to pool
	defer m.cacheWriterPool.Put(cacheWriter)

	c.Writer = cacheWriter
	c.Next()

	// only cache 2xx response
	if cacheWriter.Status() < 300 {
		cacheItem := &responseCache{}
		cacheItem.fill(cacheWriter)

		if err := options.CacheStore.Set(cacheKey, cacheItem, options.CacheDuration); err != nil {
			if options.Logger != nil {
				options.Logger.Printf("set cache error: %v", err)
			}
			return err
		}
	}

	return nil
}
