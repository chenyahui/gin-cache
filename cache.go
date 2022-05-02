package cache

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"net/url"
	"sort"
	"strings"
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

// GetCacheStrategyByRequest User can this function to design custom cache strategy by request.
// The first return value bool means whether this request should be cached.
// The second return value Strategy determine the special strategy by this request.
type GetCacheStrategyByRequest func(c *gin.Context) (bool, Strategy)

// Cache user must pass getCacheKey to describe the way to generate cache key
func Cache(
	defaultCacheStore persist.CacheStore,
	defaultExpire time.Duration,
	opts ...Option,
) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)
	return cache(defaultCacheStore, defaultExpire, cfg)
}

func cache(
	defaultCacheStore persist.CacheStore,
	defaultExpire time.Duration,
	cfg *Config,
) gin.HandlerFunc {
	if cfg.getCacheStrategyByRequest == nil {
		panic("cache strategy is nil")
	}

	sfGroup := singleflight.Group{}

	return func(c *gin.Context) {
		shouldCache, cacheStrategy := cfg.getCacheStrategyByRequest(c)
		if !shouldCache {
			c.Next()
			return
		}

		cacheKey := cacheStrategy.CacheKey

		if cfg.prefixKey != "" {
			cacheKey = cfg.prefixKey + cacheKey
		}

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
		{
			respCache := &ResponseCache{}
			err := cacheStore.Get(cacheKey, &respCache)
			if err == nil {
				replyWithCache(c, cfg, respCache)
				cfg.hitCacheCallback(c)
				return
			}

			if err != persist.ErrCacheMiss {
				cfg.logger.Errorf("get cache error: %s, cache key: %s", err, cacheKey)
			}
		}

		// cache miss, then call the backend

		// use responseCacheWriter in order to record the response
		cacheWriter := &responseCacheWriter{ResponseWriter: c.Writer}
		c.Writer = cacheWriter

		inFlight := false
		rawRespCache, _, _ := sfGroup.Do(cacheKey, func() (interface{}, error) {
			if cfg.singleFlightForgetTimeout > 0 {
				forgetTimer := time.AfterFunc(cfg.singleFlightForgetTimeout, func() {
					sfGroup.Forget(cacheKey)
				})
				defer forgetTimer.Stop()
			}

			c.Next()

			inFlight = true

			respCache := &ResponseCache{}
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
			replyWithCache(c, cfg, rawRespCache.(*ResponseCache))
			cfg.shareSingleFlightCallback(c)
		}
	}
}

// CacheByRequestURI a shortcut function for caching response by uri
func CacheByRequestURI(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	cfg := newConfigByOpts(opts...)

	var cacheStrategy GetCacheStrategyByRequest
	if cfg.ignoreQueryOrder {
		cacheStrategy = func(c *gin.Context) (bool, Strategy) {
			newUri, err := getRequestUriIgnoreQueryOrder(c.Request.RequestURI)
			if err != nil {
				cfg.logger.Errorf("getRequestUriIgnoreQueryOrder error: %s", err)
				newUri = c.Request.RequestURI
			}

			return true, Strategy{
				CacheKey: newUri,
			}
		}

	} else {
		cacheStrategy = func(c *gin.Context) (bool, Strategy) {
			return true, Strategy{
				CacheKey: c.Request.RequestURI,
			}
		}
	}

	cfg.getCacheStrategyByRequest = cacheStrategy

	return cache(defaultCacheStore, defaultExpire, cfg)
}

func getRequestUriIgnoreQueryOrder(requestURI string) (string, error) {
	parsedUrl, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return "", err
	}

	values := parsedUrl.Query()

	if len(values) == 0 {
		return requestURI, nil
	}

	queryKeys := make([]string, 0, len(values))
	for queryKey := range values {
		queryKeys = append(queryKeys, queryKey)
	}
	sort.Strings(queryKeys)

	queryVals := make([]string, 0, len(values))
	for _, queryKey := range queryKeys {
		sort.Strings(values[queryKey])
		for _, val := range values[queryKey] {
			queryVals = append(queryVals, queryKey+"="+val)
		}
	}

	return parsedUrl.Path + "?" + strings.Join(queryVals, "&"), nil
}

// CacheByRequestPath a shortcut function for caching response by url path, means will discard the query params
func CacheByRequestPath(defaultCacheStore persist.CacheStore, defaultExpire time.Duration, opts ...Option) gin.HandlerFunc {
	opts = append(opts, WithCacheStrategyByRequest(func(c *gin.Context) (bool, Strategy) {
		return true, Strategy{
			CacheKey: c.Request.URL.Path,
		}
	}))

	return Cache(defaultCacheStore, defaultExpire, opts...)
}

func init() {
	gob.Register(&ResponseCache{})
}

// ResponseCache record the http response cache
type ResponseCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func (c *ResponseCache) fillWithCacheWriter(cacheWriter *responseCacheWriter) {
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

func replyWithCache(
	c *gin.Context,
	cfg *Config,
	respCache *ResponseCache,
) {
	cfg.beforeReplyWithCacheCallback(c, respCache)

	c.Writer.WriteHeader(respCache.Status)

	for key, values := range respCache.Header {
		for _, val := range values {
			c.Writer.Header().Set(key, val)
		}
	}

	if _, err := c.Writer.Write(respCache.Data); err != nil {
		cfg.logger.Errorf("write response error: %s", err)
	}

	// abort handler chain and return directly
	c.Abort()
}
