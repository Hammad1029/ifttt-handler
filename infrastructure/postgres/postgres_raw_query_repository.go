package infrastructure

import (
	"fmt"
	"handler/common"
)

type PostgresRawQueryRepository struct {
	PostgresBaseRepository
}

func NewPostgresRawQueryRepository(base PostgresBaseRepository) *PostgresRawQueryRepository {
	return &PostgresRawQueryRepository{PostgresBaseRepository: base}
}

func (p *PostgresRawQueryRepository) RawQuery(queryString string, parameters []any) ([]common.JsonObject, error) {
	var result []common.JsonObject
	if err := p.client.Raw(queryString, parameters...).Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("method RawQuery: error running query: %s", err)
	}
	return result, nil
}

func (p *PostgresRawQueryRepository) RawSelect(queryString string, parameters []any) ([]common.JsonObject, error) {
	return p.RawQuery(queryString, parameters)
}
