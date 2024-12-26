package infrastructure

import (
	"context"
)

type MySqlRawQueryRepository struct {
	*MySqlBaseRepository
}

func NewMySqlRawQueryRepository(base *MySqlBaseRepository) *MySqlRawQueryRepository {
	return &MySqlRawQueryRepository{MySqlBaseRepository: base}
}

func (p *MySqlRawQueryRepository) RawQueryPositional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.WithContext(ctx).Raw(queryString, parameters...).Scan(&results).Error; err != nil {
		return nil, err
	}
	return &results, nil
}

func (p *MySqlRawQueryRepository) RawQueryNamed(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.WithContext(ctx).Raw(queryString, parameters).Scan(&results).Error; err != nil {
		return nil, err
	}
	return &results, nil
}

func (p *MySqlRawQueryRepository) RawExecPositional(queryString string, parameters []any, ctx context.Context) error {
	return p.client.WithContext(ctx).Exec(queryString, parameters...).Error
}

func (p *MySqlRawQueryRepository) RawExecNamed(queryString string, parameters map[string]any, ctx context.Context) error {
	return p.client.WithContext(ctx).Exec(queryString, parameters).Error
}
