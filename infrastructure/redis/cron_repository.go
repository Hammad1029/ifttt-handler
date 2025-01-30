package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"

	"github.com/redis/go-redis/v9"
)

type RedisCronRepository struct {
	*RedisBaseRepository
}

func NewRedisCronRepository(base *RedisBaseRepository) *RedisCronRepository {
	return &RedisCronRepository{RedisBaseRepository: base}
}

func (r *RedisCronRepository) StoreCrons(crons *[]api.Cron, ctx context.Context) error {
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

func (r *RedisCronRepository) GetAllCrons(ctx context.Context) (*[]api.Cron, error) {
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

func (r *RedisCronRepository) GetCronByName(name string, ctx context.Context) (*api.Cron, error) {
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
