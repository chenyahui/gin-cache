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
