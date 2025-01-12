package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	eventprofiles "ifttt/handler/domain/event_profiles"

	"github.com/redis/go-redis/v9"
)

type RedisEventProfilesRepository struct {
	*RedisBaseRepository
}

func NewRedisResponseEventRepository(base *RedisBaseRepository) *RedisEventProfilesRepository {
	return &RedisEventProfilesRepository{RedisBaseRepository: base}
}

func (r *RedisEventProfilesRepository) StoreProfiles(
	profiles *map[string]eventprofiles.Profile, ctx context.Context) error {
	if profiles == nil {
		return nil
	}
	for key, s := range *profiles {
		marshalled, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisEventProfile, key, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisEventProfilesRepository) GetProfileByTrigger(
	trigger string, ctx context.Context) (*eventprofiles.Profile, error) {
	var profile eventprofiles.Profile
	profileJSON, err := r.client.HGet(ctx, common.RedisEventProfile, trigger).Result()
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
