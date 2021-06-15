package cache

import (
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestCachePath(t *testing.T) {
	cachePathMiddleware := CacheByPath(Options{
		CacheDuration:       5 * time.Second,
		CacheStore:          persist.NewMemoryStore(1 * time.Minute),
		DisableSingleFlight: false,
	})

	_, app := gin.CreateTestContext(httptest.NewRecorder())
	app.Use(cachePathMiddleware)

	testBody := "hello world"
	app.GET("/hello", func(c *gin.Context) {
		c.String(200, testBody)
	})

	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testWriter := httptest.NewRecorder()
			app.ServeHTTP(testWriter, &http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path: "/hello",
				},
			})

			body, err := ioutil.ReadAll(testWriter.Result().Body)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, string(body), testBody)
		}()
	}

	wg.Wait()
}
