package infrastructure

import (
	"context"
	"ifttt/handler/domain/orm_schema"
)

type PostgresOrmQueryRepository struct {
	*PostgresBaseRepository
}

func NewPostgresOrmQueryRepository(base *PostgresBaseRepository) *PostgresOrmQueryRepository {
	return &PostgresOrmQueryRepository{PostgresBaseRepository: base}
}

func (m *PostgresOrmQueryRepository) ExecuteSelect(
	tableName string,
	projections map[string]string,
	populate []orm_schema.Populate,
	conditionsTemplate string,
	conditionsValue []any,
	schemaRepo orm_schema.CacheRepository,
	ctx context.Context,
) ([]map[string]any, error) {
	return nil, nil
}
