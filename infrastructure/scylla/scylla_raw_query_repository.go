package infrastructure

import (
	"fmt"
	"handler/common"

	"github.com/mitchellh/mapstructure"
)

type ScyllaRawQueryRepository struct {
	ScyllaBaseRepository
}

func NewScyllaRawQueryRepository(base ScyllaBaseRepository) *ScyllaRawQueryRepository {
	return &ScyllaRawQueryRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaRawQueryRepository) RawSelect(queryString string) ([]common.JsonObject, error) {
	rows, err := s.session.Query(queryString, nil).Iter().SliceMap()
	if err != nil {
		return nil, fmt.Errorf("method RunSelect: error running query: %s", err)
	}

	var results []common.JsonObject
	if err := mapstructure.Decode(rows, &results); err != nil {
		return nil, fmt.Errorf("method RunSelect: could not conver results to []common.JsonObject: %s", err)
	}

	return results, nil
}

func (s *ScyllaRawQueryRepository) RawSelectNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error) {
	return nil, nil
}

func (s *ScyllaRawQueryRepository) RawSelectPositional(queryString string, parameters []any) ([]common.JsonObject, error) {
	return nil, nil
}

func (s *ScyllaRawQueryRepository) RawQuery(queryString string) ([]common.JsonObject, error) {
	// query := s.session.Query(queryString, nil).Bind(parameters...)

	// if err := query.ExecRelease(); err != nil {
	// 	return nil, err
	// }

	return nil, nil
}

func (s *ScyllaRawQueryRepository) RawQueryNamed(queryString string, parameters common.JsonObject) ([]common.JsonObject, error) {
	return nil, nil
}

func (s *ScyllaRawQueryRepository) RawQueryPositional(queryString string, parameters []any) ([]common.JsonObject, error) {
	return nil, nil
}
