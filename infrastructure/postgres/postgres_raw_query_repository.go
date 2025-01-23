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

func (p *PostgresRawQueryRepository) ScanPositional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error) {
	return nil, nil
}

func (p *PostgresRawQueryRepository) ScanNamed(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error) {
	return nil, nil
}

func (p *PostgresRawQueryRepository) ExecPositional(queryString string, parameters []any, ctx context.Context) error {
	return nil
}

func (p *PostgresRawQueryRepository) ExecNamed(queryString string, parameters map[string]any, ctx context.Context) error {
	return nil
}
