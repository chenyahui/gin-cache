package persist

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisStore struct {
	RedisClient *redis.Client
}

func NewRedisStore(redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		RedisClient: redisClient,
	}
}

func (store *RedisStore) Set(key string, value interface{}, expire time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	return store.RedisClient.Set(ctx, key, string(payload), expire).Err()
}

func (store *RedisStore) Delete(key string) error {
	ctx := context.TODO()
	return store.RedisClient.Del(ctx, key).Err()
}

func (store *RedisStore) Get(key string, value interface{}) error {
	ctx := context.TODO()
	payload, err := store.RedisClient.Get(ctx, key).Bytes()

	if err == redis.Nil {
		return ErrCacheMiss
	}

	if err != nil {
		return err
	}

	return json.Unmarshal(payload, &value)
}
