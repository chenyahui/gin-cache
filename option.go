package cache

import "github.com/gin-gonic/gin"

type Config struct {
	// logger
	logger Logger

	// getCacheStrategyByRequest
	getCacheStrategyByRequest GetCacheStrategyByRequest

	// hitCacheCallback
	hitCacheCallback HitCacheCallback
}

type Option func(c *Config)

func WithLogger(l Logger) Option {
	return func(c *Config) {
		if l != nil {
			c.logger = l
		}
	}
}

func WithCacheStrategyByRequest(getGetCacheStrategyByRequest GetCacheStrategyByRequest) Option {
	return func(c *Config) {
		if getGetCacheStrategyByRequest != nil {
			c.getCacheStrategyByRequest = getGetCacheStrategyByRequest
		}
	}
}

type HitCacheCallback func(c *gin.Context)

func WithHitCacheCallback(cb HitCacheCallback) Option{
	return func(c *Config) {
		if cb != nil {
			c.hitCacheCallback = cb
		}
	}
}

var defaultHitCacheCallback = func(c *gin.Context){}

type Logger interface {
	Errorf(string, ...interface{})
}

type Discard struct {
}

func (l Discard) Errorf(string, ...interface{}) {

}
