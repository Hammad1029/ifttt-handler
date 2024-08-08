package infrastructure

import (
	"fmt"
	"handler/common"
)

type MySqlRawQueryRepository struct {
	MySqlBaseRepository
}

func NewMySqlRawQueryRepository(base MySqlBaseRepository) *MySqlRawQueryRepository {
	return &MySqlRawQueryRepository{MySqlBaseRepository: base}
}

func (p *MySqlRawQueryRepository) RawQueryPositional(queryString string, parameters []any) (*[]common.JsonObject, error) {
	var results *[]common.JsonObject
	if err := p.client.Raw(queryString, parameters...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("method *MySqlRawQueryRepository.RawQueryPositional: could not run query: %s", err)
	}
	return results, nil
}

func (p *MySqlRawQueryRepository) RawQueryNamed(queryString string, parameters common.JsonObject) (*[]common.JsonObject, error) {
	var results *[]common.JsonObject
	if err := p.client.Raw(queryString, parameters).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("method *MySqlRawQueryRepository.RawQueryNamed: could not run query: %s", err)
	}
	return results, nil
}

func (p *MySqlRawQueryRepository) RawExecPositional(queryString string, parameters []any) error {
	if err := p.client.Exec(queryString, parameters...).Error; err != nil {
		return fmt.Errorf("method *MySqlRawQueryRepository.RawExecPositional: could not run query: %s", err)
	}
	return nil
}

func (p *MySqlRawQueryRepository) RawExecNamed(queryString string, parameters common.JsonObject) error {
	if err := p.client.Exec(queryString, parameters).Error; err != nil {
		return fmt.Errorf("method *MySqlRawQueryRepository.RawExecNamed: could not run query: %s", err)
	}
	return nil
}
