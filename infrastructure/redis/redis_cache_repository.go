package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"

	"github.com/redis/go-redis/v9"
)

type RedisCacheRepository struct {
	RedisBaseRepository
}

func NewRedisCacheRepository(base RedisBaseRepository) *RedisCacheRepository {
	return &RedisCacheRepository{RedisBaseRepository: base}
}

func (r *RedisCacheRepository) StoreApis(apis *[]api.Api, ctx context.Context) error {
	if apis == nil {
		return nil
	}
	for _, api := range *apis {
		marshalled, err := json.Marshal(api)
		if err != nil {
			return fmt.Errorf("method RedisApiCacheRepository.StoreApis: could not marshall api: %s", err)
		}
		if err := r.client.HSet(ctx, common.RedisApis, api.Path, string(marshalled)).Err(); err != nil {
			return fmt.Errorf("method RedisApiCacheRepository.StoreApis: could not store api in redis: %s", err)
		}
	}
	return nil
}

func (r *RedisCacheRepository) GetAllApis(ctx context.Context) (*[]api.Api, error) {
	var apis *[]api.Api
	apiJSONs, err := r.client.HGetAll(ctx, common.RedisApis).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("method RedisApiCacheRepository.GetAllApis: could not get apis from redis: %s", err)
	}

	var apiUnmarshalled *api.Api
	for _, api := range apiJSONs {
		if err := json.Unmarshal([]byte(api), &apiUnmarshalled); err != nil {
			return nil, fmt.Errorf("method RedisApiCacheRepository.GetAllApis: could not unmarshall api: %s", err)
		}
		*apis = append(*apis, *apiUnmarshalled)
	}

	return apis, nil
}

func (r *RedisCacheRepository) GetApiByPath(path string, ctx context.Context) (*api.Api, error) {
	var api *api.Api
	apiJSON, err := r.client.HGet(ctx, common.RedisApis, path).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("method RedisApiCacheRepository.GetApiByGroupAndName: error in getting api: %s", err)
	}
	if err := json.Unmarshal([]byte(apiJSON), &api); err != nil {
		return nil, fmt.Errorf("method RedisApiCacheRepository.GetApiByGroupAndName: error in unmarshalling api: %s", err)
	}
	return api, nil
}

func (r *RedisCacheRepository) StoreCrons(crons *[]api.Cron, ctx context.Context) error {
	if crons == nil {
		return nil
	}
	for _, c := range *crons {
		marshalled, err := json.Marshal(c)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisCrons, c.Name, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisCacheRepository) GetAllCrons(ctx context.Context) (*[]api.Cron, error) {
	var crons *[]api.Cron
	cronJSONs, err := r.client.HGetAll(ctx, common.RedisCrons).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cronUnmarshalled *api.Cron
	for _, api := range cronJSONs {
		if err := json.Unmarshal([]byte(api), &cronUnmarshalled); err != nil {
			return nil, err
		}
		*crons = append(*crons, *cronUnmarshalled)
	}

	return crons, nil
}

func (r *RedisCacheRepository) GetCronByName(name string, ctx context.Context) (*api.Cron, error) {
	var cron *api.Cron
	cronJSON, err := r.client.HGet(ctx, common.RedisCrons, name).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(cronJSON), &cron); err != nil {
		return nil, err
	}
	return cron, nil
}
