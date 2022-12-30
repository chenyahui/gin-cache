package main

import (
	"fmt"
	"sync/atomic"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.New()

	memoryStore := persist.NewMemoryStore(1 * time.Minute)

	var cacheHitCount, cacheMissCount int32
	app.GET("/hello",
		cache.CacheByRequestURI(
			memoryStore,
			2*time.Second,
			cache.WithOnHitCache(func(c *gin.Context) {
				atomic.AddInt32(&cacheHitCount, 1)
			}),
			cache.WithOnMissCache(func(c *gin.Context) {
				atomic.AddInt32(&cacheMissCount, 1)
			}),
		),
		func(c *gin.Context) {
			c.String(200, "hello world")
		},
	)

	app.GET("/get_hit_count", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("total hit count: %d", cacheHitCount))
	})
	app.GET("/get_miss_count", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("total miss count: %d", cacheMissCount))
	})

	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
