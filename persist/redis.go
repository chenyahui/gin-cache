package persist

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
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
	payload, err := Serialize(value)
	if err != nil {
		return err
	}

	ctx := context.TODO()
	return store.RedisClient.Set(ctx, key, payload, expire).Err()
}

func (store *RedisStore) Delete(key string) error {
	ctx := context.TODO()
	return store.RedisClient.Del(ctx, key).Err()
}

func (store *RedisStore) Get(key string, value interface{}) error {
	ctx := context.TODO()
	payload, err := store.RedisClient.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return ErrCacheMiss
	}

	if err != nil {
		return err
	}

	return Deserialize(payload, value)
}
