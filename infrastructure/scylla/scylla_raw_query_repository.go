package infrastructure

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/scylladb/gocqlx/v3"
)

type ScyllaRawQueryRepository struct {
	ScyllaBaseRepository
}

func NewScyllaRawQueryRepository(base ScyllaBaseRepository) *ScyllaRawQueryRepository {
	return &ScyllaRawQueryRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaRawQueryRepository) RawQueryPositional(queryString string, parameters []any) (*[]map[string]any, error) {
	var results *[]map[string]any
	query := s.session.Query(queryString, nil)
	defer query.Release()
	if rows, err := query.Bind(parameters...).Iter().SliceMap(); err != nil {
		return nil, fmt.Errorf("method *ScyllaRawQueryRepository.RawQueryPositional: could not get slice map: %s", err)
	} else {
		if err := mapstructure.Decode(rows, &results); err != nil {
			return nil, fmt.Errorf("method *ScyllaRawQueryRepository.RawQueryPositional: could not decode query results: %s", err)
		}
	}

	return results, nil
}

func (s *ScyllaRawQueryRepository) RawQueryNamed(queryString string, parameters map[string]any) (*[]map[string]any, error) {
	var results *[]map[string]any
	stmt, names, err := gocqlx.CompileNamedQueryString(queryString)
	if err != nil {
		return nil, fmt.Errorf("method *ScyllaRawQueryRepository.RawQueryNamed: could not compile named query string: %s", err)
	}
	query := s.session.Query(stmt, names)
	defer query.Release()
	if rows, err := query.BindMap(parameters).Iter().SliceMap(); err != nil {
		return nil, fmt.Errorf("method *ScyllaRawQueryRepository.RawQueryNamed: could not get slice map: %s", err)
	} else {
		if err := mapstructure.Decode(rows, &results); err != nil {
			return nil, fmt.Errorf("method *ScyllaRawQueryRepository.RawQueryNamed: could not decode query results: %s", err)
		}
	}

	return results, nil
}

func (s *ScyllaRawQueryRepository) RawExecPositional(queryString string, parameters []any) error {
	query := s.session.Query(queryString, nil)
	if err := query.Bind(parameters...).ExecRelease(); err != nil {
		return fmt.Errorf("method *ScyllaRawQueryRepository.RawExecPositional: could not get slice map: %s", err)
	}
	return nil
}

func (s *ScyllaRawQueryRepository) RawExecNamed(queryString string, parameters map[string]any) error {
	stmt, names, err := gocqlx.CompileNamedQueryString(queryString)
	if err != nil {
		return fmt.Errorf("method *ScyllaRawQueryRepository.RawExecNamed: could not compile named query string: %s", err)
	}
	query := s.session.Query(stmt, names)
	if err := query.BindMap(parameters).ExecRelease(); err != nil {
		return fmt.Errorf("method *ScyllaRawQueryRepository.RawExecNamed: could not get slice map: %s", err)
	}
	return nil
}
