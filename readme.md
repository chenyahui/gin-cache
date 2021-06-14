# gin-cache
A high performance gin middleware to cache http response. Compared to gin-contrib/cache, it has more than 45% performance improvement.

# How To Use

## Install
> go get github.com/chenyahui/gin-cache

## Example
```go
package main

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.New()

	app.GET("/hello",
		cache.CacheByPath(cache.Options{
			CacheDuration:       5 * time.Second,
			CacheStore:          persist.NewMemoryStore(1 * time.Minute),
			DisableSingleFlight: true,
		}),
		func(c *gin.Context) {
			time.Sleep(200 * time.Millisecond)
			c.String(200, "hello world")
		},
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
```

# Feature
* Has more than 45% performance improvement compared to gin-cache
* Offer a way to custom the cache key of request
* Use Sync.Pool to cache high frequency objects
* Use SingleFlight to avoid Hotspot Invalid

# Benchmark
```
wrk -c 500 -d 1m -t 5 http://127.0.0.1:8080/hello
```

![QPS](https://www.cyhone.com/img/gin-cache/qps.png)
