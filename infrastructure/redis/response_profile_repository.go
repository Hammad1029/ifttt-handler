package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	"ifttt/handler/domain/configuration"

	"github.com/redis/go-redis/v9"
)

type RedisResponseProfilesRepository struct {
	*RedisBaseRepository
}

func NewRedisResponseProfilesRepository(base *RedisBaseRepository) *RedisResponseProfilesRepository {
	return &RedisResponseProfilesRepository{RedisBaseRepository: base}
}

func (r *RedisResponseProfilesRepository) StoreProfiles(
	profiles *map[string]configuration.ResponseProfile, ctx context.Context) error {
	if profiles == nil {
		return nil
	}
	for key, s := range *profiles {
		marshalled, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisResponseProfile, key, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisResponseProfilesRepository) GetProfileByName(
	name string, ctx context.Context) (*configuration.ResponseProfile, error) {
	var profile configuration.ResponseProfile
	profileJSON, err := r.client.HGet(ctx, common.RedisResponseProfile, name).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal([]byte(profileJSON), &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}
