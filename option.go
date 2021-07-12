package cache

type Config struct {
	// logger
	logger Logger

	// getCacheStrategyByRequest
	getCacheStrategyByRequest GetCacheStrategyByRequest
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

type Logger interface {
	Errorf(string, ...interface{})
}

type Discard struct {
}

func (l Discard) Errorf(string, ...interface{}) {

}
