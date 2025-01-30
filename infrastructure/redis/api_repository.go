package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"

	"github.com/redis/go-redis/v9"
)

type RedisAPIRepository struct {
	*RedisBaseRepository
}

func NewRedisAPIRepository(base *RedisBaseRepository) *RedisAPIRepository {
	return &RedisAPIRepository{RedisBaseRepository: base}
}

func (r *RedisAPIRepository) StoreApis(apis *[]api.Api, ctx context.Context) error {
	if apis == nil {
		return nil
	}
	for _, api := range *apis {
		marshalled, err := json.Marshal(api)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisApis, api.Path, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisAPIRepository) GetAllApis(ctx context.Context) (*[]api.Api, error) {
	var apis []api.Api
	apiJSONs, err := r.client.HGetAll(ctx, common.RedisApis).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var apiUnmarshalled api.Api
	for _, api := range apiJSONs {
		if err := json.Unmarshal([]byte(api), &apiUnmarshalled); err != nil {
			return nil, err
		}
		apis = append(apis, apiUnmarshalled)
	}

	return &apis, nil
}

func (r *RedisAPIRepository) GetApiByPath(path string, ctx context.Context) (*api.Api, error) {
	var api api.Api
	apiJSON, err := r.client.HGet(ctx, common.RedisApis, path).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(apiJSON), &api); err != nil {
		return nil, err
	}
	return &api, nil
}
