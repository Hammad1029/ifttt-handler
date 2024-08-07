package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/domain/api"

	"github.com/redis/go-redis/v9"
)

type RedisApiCacheRepository struct {
	RedisBaseRepository
}

func NewRedisApiCacheRepository(base RedisBaseRepository) *RedisApiCacheRepository {
	return &RedisApiCacheRepository{RedisBaseRepository: base}
}

func (r *RedisApiCacheRepository) StoreApis(apis []api.Api, ctx context.Context) error {
	for _, api := range apis {
		marshalled, err := json.Marshal(api)
		if err != nil {
			return fmt.Errorf("method RedisApiCacheRepository.StoreApis: could not marshall api: %s", err)
		}
		if err := r.client.HSet(ctx, "apis", fmt.Sprintf("%s.%s", api.ApiGroup, api.ApiName), string(marshalled)).Err(); err != nil {
			return fmt.Errorf("method RedisApiCacheRepository.StoreApis: could not store api in redis: %s", err)
		}
	}
	return nil
}

func (r *RedisApiCacheRepository) GetAllApis(ctx context.Context) ([]api.Api, error) {
	var apis []api.Api
	apiJSONs, err := r.client.HGetAll(context.Background(), "apis").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("method RedisApiCacheRepository.GetAllApis: could not get apis from redis: %s", err)
	}

	var apiUnmarshalled api.Api
	for _, api := range apiJSONs {
		if err := json.Unmarshal([]byte(api), &apiUnmarshalled); err != nil {
			return nil, fmt.Errorf("method RedisApiCacheRepository.GetAllApis: could not unmarshall api: %s", err)
		}
		apis = append(apis, apiUnmarshalled)
	}

	return apis, nil
}

func (r *RedisApiCacheRepository) GetApiByGroupAndName(group string, name string, ctx context.Context) (api.Api, error) {
	var api api.Api
	apiJSON, err := r.client.HGet(ctx, "apis", fmt.Sprintf("%s.%s", group, name)).Result()
	if err == redis.Nil {
		return api, fmt.Errorf("method RedisApiCacheRepository.GetApiByGroupAndName: no api found with group: %s and name: %s", group, name)
	}
	if err != nil {
		return api, fmt.Errorf("method RedisApiCacheRepository.GetApiByGroupAndName: error in getting api: %s", err)
	}
	err = json.Unmarshal([]byte(apiJSON), &api)
	if err != nil {
		return api, fmt.Errorf("method RedisApiCacheRepository.GetApiByGroupAndName: error in unmarshalling api: %s", err)
	}
	return api, nil
}
