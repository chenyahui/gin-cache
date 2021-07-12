package cache

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"sync"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

// Strategy the cache strategy
type Strategy struct {
	CacheKey string

	// CacheStore if nil, use default cache store instead
	CacheStore persist.CacheStore

	// CacheDuration
	CacheDuration time.Duration
}

type GetCacheStrategyByRequest func(c *gin.Context) (bool, Strategy)

// Cache user must pass getCacheKey to describe the way to generate cache key
func Cache(
	defaultCacheStore persist.CacheStore,
	defaultExpire time.Duration,
	opts ...Option,
) gin.HandlerFunc {
	cfg := &Config{
		logger: Discard{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	sfGroup := singleflight.Group{}
	respCachePool := newResponseCachePool()

	return func(c *gin.Context) {
		shouldCache, cacheStrategy := cfg.getCacheStrategyByRequest(c)
		if !shouldCache {
			c.Next()
			return
		}

		cacheKey := cacheStrategy.CacheKey

		// merge cfg
		cacheStore := defaultCacheStore
		if cacheStrategy.CacheStore != nil {
			cacheStore = cacheStrategy.CacheStore
		}

		cacheDuration := defaultExpire
		if cacheStrategy.CacheDuration > 0 {
			cacheDuration = cacheStrategy.CacheDuration
		}

		// read cache first
		respCache := &responseCache{}

		err := cacheStore.Get(cacheKey, &respCache)
		if err == nil {
			replyWithCache(c, respCache)
			respCachePool.Put(respCache)
			return
		}

		if err != persist.ErrCacheMiss {
			cfg.logger.Errorf("get cache error: %s, cache key: %s", err, cacheKey)
		}

		// use responseCacheWriter in order to record the response
		cacheWriter := &responseCacheWriter{ResponseWriter: c.Writer}
		c.Writer = cacheWriter

		inFlight := false
		rawRespCache, _, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
			c.Next()

			inFlight = true

			respCache.fillWithCacheWriter(cacheWriter)

			// only cache 2xx response
			if !c.IsAborted() && cacheWriter.Status() < 300 && cacheWriter.Status() >= 200 {
				if err := cacheStore.Set(cacheKey, respCache, cacheDuration); err != nil {
					cfg.logger.Errorf("set cache key error: %s, cache key: %s", err, cacheKey)
				}
			}

			return respCache, nil
		})

		if !inFlight {
			replyWithCache(c, rawRespCache.(*responseCache))
		}
	}
}

// CacheByRequestURI a shortcut function for caching response with uri
func CacheByRequestURI(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	opts = append(opts, WithCacheStrategyByRequest(func(c *gin.Context) (bool, Strategy) {
		return true, Strategy{
			CacheKey: c.Request.RequestURI,
		}
	}))
	return Cache(defaultCacheStore, defaultExpire, opts...)
}

// CacheByRequestPath a shortcut function for caching response with url path, discard the query params
func CacheByRequestPath(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	opts = append(opts, WithCacheStrategyByRequest(func(c *gin.Context) (bool, Strategy) {
		return true, Strategy{
			CacheKey: c.Request.URL.Path,
		}
	}))

	return Cache(defaultCacheStore, defaultExpire, opts...)
}

func init() {
	gob.Register(&responseCache{})
}

type responseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func newResponseCache() *responseCache {
	return &responseCache{
		Status: 0,
		Header: make(http.Header),
	}
}

func (c *responseCache) reset() {
	c.Data = c.Data[0:0]
	c.Header = make(http.Header)
	c.Status = 0
}

func (c *responseCache) fillWithCacheWriter(cacheWriter *responseCacheWriter) {
	c.Status = cacheWriter.Status()
	c.Data = cacheWriter.body.Bytes()
	c.Header = cacheWriter.Header().Clone()
}

// responseCacheWriter
type responseCacheWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *responseCacheWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseCacheWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *responseCacheWriter) reset(writer gin.ResponseWriter) {
	w.body.Reset()
	w.ResponseWriter = writer
}

type responseCachePool struct {
	pool *sync.Pool
}

func newResponseCachePool() *responseCachePool {
	return &responseCachePool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &responseCache{
					Header: make(http.Header),
				}
			},
		},
	}
}

func (p *responseCachePool) Get() *responseCache {
	respCache := p.pool.Get().(*responseCache)
	respCache.reset()

	return respCache
}

func (p *responseCachePool) Put(c *responseCache) {
	p.pool.Put(c)
}

func replyWithCache(c *gin.Context, respCache *responseCache) {
	c.Writer.WriteHeader(respCache.Status)
	for key, values := range respCache.Header {
		for _, val := range values {
			c.Writer.Header().Add(key, val)
		}
	}

	if _, err := c.Writer.Write(respCache.Data); err != nil {
		//logrus.Errorf("write response error: %s", err)
	}

	// abort handler chain and return directly
	c.Abort()
}
