package infrastructure

import (
	"context"
	"fmt"
	redisInfra "ifttt/handler/infrastructure/redis"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

const redisCache = "redis"

type RedisStore struct {
	client *redis.Client
	config redisConfig
}

type redisConfig struct {
	Db       string `json:"db" mapstructure:"db"`
	Host     string `json:"host" mapstructure:"host"`
	Port     string `json:"port" mapstructure:"port"`
	Password string `json:"password" mapstructure:"password"`
	DbIndex  string `json:"dbIndex" mapstructure:"dbIndex"`
}

func (r *RedisStore) init(config map[string]any) error {
	if err := mapstructure.Decode(config, &r.config); err != nil {
		return fmt.Errorf("method: *RedisStore.Init: could not decode redis configuration from env: %s", err)
	}
	dbIndex, _ := strconv.Atoi(r.config.DbIndex)
	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", r.config.Host, r.config.Port),
		Password: r.config.Password,
		DB:       dbIndex,
	})
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("method: *RedisStore.Init: could not connect to redis: %s", err)
	}
	return nil
}

func (r *RedisStore) createCacheStore() *CacheStore {
	redisBase := redisInfra.NewRedisBaseRepository(r.client)
	return &CacheStore{
		Store:    r,
		APIRepo:  redisInfra.NewRedisAPIRepository(redisBase),
		CronRepo: redisInfra.NewRedisCronRepository(redisBase),
		OrmRepo:  redisInfra.NewRedisOrmSchemaRepository(redisBase),
	}
}

func (r *RedisStore) createAppCacheStore() *AppCacheStore {
	redisBase := redisInfra.NewRedisBaseRepository(r.client)
	return &AppCacheStore{
		Store:        r,
		AppCacheRepo: redisInfra.NewRedisAppCacheRepository(redisBase),
	}
}
