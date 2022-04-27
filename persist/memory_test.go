package persist

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	memoryStore := NewMemoryStore(1 * time.Minute)

	expectVal := "123"
	require.Nil(t, memoryStore.Set("test", expectVal, 1*time.Second))

	value := ""
	assert.Nil(t, memoryStore.Get("test", &value))
	assert.Equal(t, expectVal, value)

	time.Sleep(1 * time.Second)
	assert.Equal(t, ErrCacheMiss, memoryStore.Get("test", &value))
}
