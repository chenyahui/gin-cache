# gin-cache
[![Release](https://img.shields.io/github/release/chenyahui/gin-cache.svg?style=flat-square)](https://github.com/chenyahui/gin-cache/releases)
[![doc](https://img.shields.io/badge/go.dev-doc-007d9c?style=flat-square&logo=read-the-docs)](https://pkg.go.dev/github.com/chenyahui/gin-cache)
[![goreportcard for gin-cache](https://goreportcard.com/badge/github.com/chenyahui/gin-cache)](https://goreportcard.com/report/github.com/chenyahui/gin-cache)
![](https://img.shields.io/badge/license-MIT-green)
[![codecov](https://codecov.io/gh/chenyahui/gin-cache/branch/main/graph/badge.svg?token=MX8Z4D5RZS)](https://codecov.io/gh/chenyahui/gin-cache)

English | [🇨🇳中文](README_ZH.md)

A high performance gin middleware to cache http response. Compared to gin-contrib/cache. It has a huge performance improvement.


# Feature

* Has a huge performance improvement compared to gin-contrib/cache.
* Cache http response in local memory or Redis.
* Offer a way to custom the cache strategy by per request.
* Use singleflight to avoid cache breakdown problem.
* Only Cache 2xx HTTP Response.

# How To Use

## Install
```
go get -u github.com/chenyahui/gin-cache
```

## Example

### Cache In Local Memory

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

	memoryStore := persist.NewMemoryStore(1 * time.Minute)

	app.GET("/hello",
		cache.CacheByRequestURI(memoryStore, 2*time.Second),
		func(c *gin.Context) {
			c.String(200, "hello world")
		},
	)

	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
```

### Cache In Redis

```go
package main

import (
	"time"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	app := gin.New()

	redisStore := persist.NewRedisStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	}))

	app.GET("/hello",
		cache.CacheByRequestURI(redisStore, 2*time.Second),
		func(c *gin.Context) {
			c.String(200, "hello world")
		},
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
```



# Benchmark

```
wrk -c 500 -d 1m -t 5 http://127.0.0.1:8080/hello
```

## MemoryStore

![MemoryStore QPS](https://www.cyhone.com/img/gin-cache/memory_cache_qps.png)

## RedisStore

![RedisStore QPS](https://www.cyhone.com/img/gin-cache/redis_cache_qps.png)
