package infrastructure

import (
	"fmt"
)

type PostgresRawQueryRepository struct {
	*PostgresBaseRepository
}

func NewPostgresRawQueryRepository(base *PostgresBaseRepository) *PostgresRawQueryRepository {
	return &PostgresRawQueryRepository{PostgresBaseRepository: base}
}

func (p *PostgresRawQueryRepository) RawQueryPositional(queryString string, parameters []any) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.Raw(queryString, parameters...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("method *PostgresRawQueryRepository.RawQueryPositional: could not run query: %s", err)
	}
	return &results, nil
}

func (p *PostgresRawQueryRepository) RawQueryNamed(queryString string, parameters map[string]any) (*[]map[string]any, error) {
	var results []map[string]any
	if err := p.client.Raw(queryString, parameters).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("method *PostgresRawQueryRepository.RawQueryNamed: could not run query: %s", err)
	}
	return &results, nil
}

func (p *PostgresRawQueryRepository) RawExecPositional(queryString string, parameters []any) error {
	if err := p.client.Exec(queryString, parameters...).Error; err != nil {
		return fmt.Errorf("method *PostgresRawQueryRepository.RawExecPositional: could not run query: %s", err)
	}
	return nil
}

func (p *PostgresRawQueryRepository) RawExecNamed(queryString string, parameters map[string]any) error {
	if err := p.client.Exec(queryString, parameters).Error; err != nil {
		return fmt.Errorf("method *PostgresRawQueryRepository.RawExecNamed: could not run query: %s", err)
	}
	return nil
}
