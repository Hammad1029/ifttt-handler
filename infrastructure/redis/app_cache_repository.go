package infrastructure

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAppCacheRepository struct {
	*RedisBaseRepository
}

func NewRedisAppCacheRepository(base *RedisBaseRepository) *RedisAppCacheRepository {
	return &RedisAppCacheRepository{RedisBaseRepository: base}
}

func (r *RedisAppCacheRepository) SetKey(key string, val any, ttl uint, ctx context.Context) error {
	if err := r.client.Set(ctx, key, val, time.Duration(ttl)*time.Second).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisAppCacheRepository) GetKey(key string, ctx context.Context) (any, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (r *RedisAppCacheRepository) DeleteKey(key string, ctx context.Context) (int64, error) {
	affected, err := r.client.Del(ctx, key).Result()
	return affected, err
}
