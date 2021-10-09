package cache

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Config contains all options
type Config struct {
	logger Logger

	getCacheStrategyByRequest GetCacheStrategyByRequest

	hitCacheCallback OnHitCacheCallback

	singleFlightForgetTimeout time.Duration
	shareSingleFlightCallback OnShareSingleFlightCallback
}

// Option represents the optional function.
type Option func(c *Config)

// WithLogger set the custom logger
func WithLogger(l Logger) Option {
	return func(c *Config) {
		if l != nil {
			c.logger = l
		}
	}
}

// Logger define the logger interface
type Logger interface {
	Errorf(string, ...interface{})
}

// Discard the default logger that will discard all logs of gin-cache
type Discard struct {
}

// Errorf will output the log at error level
func (l Discard) Errorf(string, ...interface{}) {

}

// WithCacheStrategyByRequest set up the custom strategy by per request
func WithCacheStrategyByRequest(getGetCacheStrategyByRequest GetCacheStrategyByRequest) Option {
	return func(c *Config) {
		if getGetCacheStrategyByRequest != nil {
			c.getCacheStrategyByRequest = getGetCacheStrategyByRequest
		}
	}
}

// OnHitCacheCallback define the callback when use cache
type OnHitCacheCallback func(c *gin.Context)

var defaultHitCacheCallback = func(c *gin.Context) {}

// WithOnHitCache will be called when cache hit.
func WithOnHitCache(cb OnHitCacheCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.hitCacheCallback = cb
		}
	}
}

// OnShareSingleFlightCallback define the callback when share the singleflight result
type OnShareSingleFlightCallback func(c *gin.Context)

var defaultShareSingleFlightCallback = func(c *gin.Context) {}

// WithOnShareSingleFlight will be called when share the singleflight result
func WithOnShareSingleFlight(cb OnShareSingleFlightCallback) Option {
	return func(c *Config) {
		if cb != nil {
			c.shareSingleFlightCallback = cb
		}
	}
}

// WithSingleFlightForgetTimeout to reduce the impact of long tail requests. when request in the singleflight,
// after the forget timeout, singleflight.Forget will be called
func WithSingleFlightForgetTimeout(forgetTimeout time.Duration) Option {
	return func(c *Config) {
		if forgetTimeout > 0 {
			c.singleFlightForgetTimeout = forgetTimeout
		}
	}
}
