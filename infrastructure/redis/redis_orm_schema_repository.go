package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
)

type RedisOrmRepository struct {
	*RedisBaseRepository
}

func NewRedisOrmSchemaRepository(base *RedisBaseRepository) *RedisOrmRepository {
	return &RedisOrmRepository{RedisBaseRepository: base}
}

func (r *RedisOrmRepository) GetModel(name string, ctx context.Context) (*orm_schema.Model, error) {
	var schema orm_schema.Model
	scehmaJSON, err := r.client.HGet(ctx, common.RedisSchemas, name).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(scehmaJSON), &schema); err != nil {
		return nil, err
	}
	return &schema, nil

}

func (r *RedisOrmRepository) SetModels(schemas *map[string]orm_schema.Model, ctx context.Context) error {
	if schemas == nil {
		return nil
	}
	for key, s := range *schemas {
		marshalled, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisSchemas, key, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisOrmRepository) GetAssociation(name string, ctx context.Context) (*orm_schema.ModelAssociation, error) {
	var association orm_schema.ModelAssociation
	associationJSON, err := r.client.HGet(ctx, common.RedisAssociatons, name).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(associationJSON), &association); err != nil {
		return nil, err
	}
	return &association, nil
}

func (r *RedisOrmRepository) SetAssociations(associations *map[string]orm_schema.ModelAssociation, ctx context.Context) error {
	if associations == nil {
		return nil
	}
	for key, s := range *associations {
		marshalled, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisAssociatons, key, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil
}
