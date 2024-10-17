package infrastructure

import (
	"context"
)

type PostgresRawQueryRepository struct {
	*PostgresBaseRepository
}

func NewPostgresRawQueryRepository(base *PostgresBaseRepository) *PostgresRawQueryRepository {
	return &PostgresRawQueryRepository{PostgresBaseRepository: base}
}

func (p *PostgresRawQueryRepository) RawQueryPositional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.WithContext(ctx).Raw(queryString, parameters...).Scan(&results).Error; err != nil {
		return nil, err
	}
	return &results, nil
}

func (p *PostgresRawQueryRepository) RawQueryNamed(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.WithContext(ctx).Raw(queryString, parameters).Scan(&results).Error; err != nil {
		return nil, err
	}
	return &results, nil
}

func (p *PostgresRawQueryRepository) RawExecPositional(queryString string, parameters []any, ctx context.Context) error {
	return p.client.WithContext(ctx).Exec(queryString, parameters...).Error
}

func (p *PostgresRawQueryRepository) RawExecNamed(queryString string, parameters map[string]any, ctx context.Context) error {
	return p.client.WithContext(ctx).Exec(queryString, parameters).Error
}
