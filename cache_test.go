package cache

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func mockHttpRequest(middleware gin.HandlerFunc, url string, withRand bool) *httptest.ResponseRecorder {
	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)
	engine.Use(middleware)
	engine.GET("/cache", func(c *gin.Context) {
		body := "uid:" + c.Query("uid")
		if withRand {
			body += fmt.Sprintf(",rand:%d", rand.Int())
		}
		c.String(http.StatusOK, body)
	})

	testRequest := httptest.NewRequest(http.MethodGet, url, nil)

	engine.ServeHTTP(testWriter, testRequest)

	return testWriter
}

func TestCacheByRequestPath(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cachePathMiddleware := CacheByRequestPath(memoryStore, 3*time.Second)

	w1 := mockHttpRequest(cachePathMiddleware, "/cache?uid=u1", true)
	w2 := mockHttpRequest(cachePathMiddleware, "/cache?uid=u2", true)
	w3 := mockHttpRequest(cachePathMiddleware, "/cache?uid=u3", true)

	assert.NotEqual(t, w1.Body.String(), "")
	assert.Equal(t, w1.Body.String(), w2.Body.String())
	assert.Equal(t, w2.Body.String(), w3.Body.String())
	assert.Equal(t, w1.Code, w2.Code)
}

func TestCacheDuration(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 3*time.Second)

	w1 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1", true)
	time.Sleep(1 * time.Second)

	w2 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1", true)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
	assert.Equal(t, w1.Code, w2.Code)
	time.Sleep(2 * time.Second)

	w3 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1", true)
	assert.NotEqual(t, w1.Body.String(), w3.Body.String())
}

func TestCacheByRequestURI(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 3*time.Second)

	w1 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1", true)
	w2 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1", true)
	w3 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u2", true)

	assert.Equal(t, w1.Body.String(), w2.Body.String())
	assert.Equal(t, w1.Code, w2.Code)

	assert.NotEqual(t, w2.Body.String(), w3.Body.String())

	w4 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u4", false)
	assert.Equal(t, "uid:u4", w4.Body.String())
}

func TestHeader(t *testing.T) {
	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)

	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 3*time.Second)

	engine.Use(func(c *gin.Context) {
		c.Header("test_header_key", "test_header_value")
	})

	engine.Use(cacheURIMiddleware)

	engine.GET("/cache", func(c *gin.Context) {
		c.Header("test_header_key", "test_header_value2")
		c.String(http.StatusOK, "value")
	})

	testRequest := httptest.NewRequest(http.MethodGet, "/cache", nil)

	{
		engine.ServeHTTP(testWriter, testRequest)
		values := testWriter.Header().Values("test_header_key")
		assert.Equal(t, 1, len(values))
		assert.Equal(t, "test_header_value2", values[0])

	}

	{
		engine.ServeHTTP(testWriter, testRequest)
		values := testWriter.Header().Values("test_header_key")
		assert.Equal(t, 1, len(values))
		assert.Equal(t, "test_header_value2", values[0])
	}
}

func TestConcurrentRequest(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 1*time.Second)

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			uid := rand.Intn(5)
			url := fmt.Sprintf("/cache?uid=%d", uid)
			expect := fmt.Sprintf("uid:%d", uid)

			writer := mockHttpRequest(cacheURIMiddleware, url, false)
			assert.Equal(t, expect, writer.Body.String())
		}()
	}

	wg.Wait()
}

func TestWriteHeader(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 1*time.Second)

	testWriter := httptest.NewRecorder()

	_, engine := gin.CreateTestContext(testWriter)
	engine.Use(cacheURIMiddleware)
	engine.GET("/cache", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Header().Set("hello", "world")
	})

	{
		testRequest := httptest.NewRequest(http.MethodGet, "/cache", nil)
		engine.ServeHTTP(testWriter, testRequest)
		assert.Equal(t, "world", testWriter.Header().Get("hello"))
	}

	{
		testRequest := httptest.NewRequest(http.MethodGet, "/cache", nil)
		engine.ServeHTTP(testWriter, testRequest)
		assert.Equal(t, "world", testWriter.Header().Get("hello"))
	}
}

func TestGetRequestUriIgnoreQueryOrder(t *testing.T) {
	val, err := getRequestUriIgnoreQueryOrder("/test?c=3&b=2&a=1")
	require.Nil(t, err)
	assert.Equal(t, "/test?a=1&b=2&c=3", val)

	val, err = getRequestUriIgnoreQueryOrder("/test?d=4&e=5")
	require.Nil(t, err)
	assert.Equal(t, "/test?d=4&e=5", val)
}

func TestCacheByRequestURIIgnoreOrder(t *testing.T) {
	memoryStore := persist.NewMemoryStore(1 * time.Minute)
	cacheURIMiddleware := CacheByRequestURI(memoryStore, 3*time.Second, IgnoreQueryOrder())

	w1 := mockHttpRequest(cacheURIMiddleware, "/cache?uid=u1&a=2", true)
	w2 := mockHttpRequest(cacheURIMiddleware, "/cache?a=2&uid=u1", true)

	assert.Equal(t, w1.Body.String(), w2.Body.String())
	assert.Equal(t, w1.Code, w2.Code)
}
