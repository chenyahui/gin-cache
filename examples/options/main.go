package main

import (
	"fmt"
	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"sync/atomic"
	"time"
)

func main() {
	app := gin.New()

	memoryStore := persist.NewMemoryStore(1 * time.Minute)

	var cacheHitCount int32
	app.GET("/hello",
		cache.CacheByRequestURI(
			memoryStore,
			2*time.Second,
			cache.WithOnHitCache(func(c *gin.Context) {
				atomic.AddInt32(&cacheHitCount, 1)
			}),
		),
		func(c *gin.Context) {
			c.String(200, "hello world")
		},
	)

	app.GET("/get_hit_count", func(c *gin.Context) {
		c.String(200, fmt.Sprintf("total hit count: %d", cacheHitCount))
	})

	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
