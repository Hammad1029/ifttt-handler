package infrastructure

import (
	"handler/common"
)

type PostgresRawQueryRepository struct {
	PostgresBaseRepository
}

func NewPostgresRawQueryRepository(base PostgresBaseRepository) *PostgresRawQueryRepository {
	return &PostgresRawQueryRepository{PostgresBaseRepository: base}
}

func (p *PostgresRawQueryRepository) RawSelect(queryString string) ([]common.JsonObject, error) {
	// return p.RawQuery(queryString, parameters)
	return nil, nil
}

func (s *PostgresRawQueryRepository) RawSelectNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error) {
	return nil, nil
}

func (s *PostgresRawQueryRepository) RawSelectPositional(queryString string, parameters []any) ([]common.JsonObject, error) {
	return nil, nil
}

func (p *PostgresRawQueryRepository) RawQuery(queryString string) ([]common.JsonObject, error) {
	// var result []common.JsonObject
	// if err := p.client.Raw(queryString, parameters...).Scan(&result).Error; err != nil {
	// 	return nil, fmt.Errorf("method RawQuery: error running query: %s", err)
	// }
	// return result, nil
	return nil, nil
}

func (s *PostgresRawQueryRepository) RawQueryNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error) {
	return nil, nil
}

func (s *PostgresRawQueryRepository) RawQueryPositional(queryString string, parameters []any) ([]common.JsonObject, error) {
	return nil, nil
}
