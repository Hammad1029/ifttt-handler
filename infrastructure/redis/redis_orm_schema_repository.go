package infrastructure

import (
	"context"
	"encoding/json"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
)

type RedisOrmSchemaRepository struct {
	*RedisBaseRepository
}

func NewRedisOrmSchemaRepository(base *RedisBaseRepository) *RedisOrmSchemaRepository {
	return &RedisOrmSchemaRepository{RedisBaseRepository: base}
}

func (r *RedisOrmSchemaRepository) GetSchema(tableName string, ctx context.Context) (*orm_schema.Schema, error) {
	var schema orm_schema.Schema
	scehmaJSON, err := r.client.HGet(ctx, common.RedisSchemas, tableName).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(scehmaJSON), &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

func (r *RedisOrmSchemaRepository) StoreSchema(schemas *[]orm_schema.Schema, ctx context.Context) error {
	if schemas == nil {
		return nil
	}
	for _, s := range *schemas {
		marshalled, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := r.client.HSet(ctx, common.RedisSchemas, s.TableName, string(marshalled)).Err(); err != nil {
			return err
		}
	}
	return nil

}
