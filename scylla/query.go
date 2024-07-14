package scylla

import (
	"fmt"

	jsontocql "github.com/Hammad1029/json-to-cql"
)

func RunSelect(pQuery jsontocql.ParameterizedQuery, parameters []interface{}) ([]map[string]interface{}, error) {
	queryString, err := pQuery.ResolveQuery(parameters)
	if err != nil {
		return nil, fmt.Errorf("method RunSelect: error resolving query: %s", err)
	}

	rows, err := GetScylla().Query(queryString, nil).Iter().SliceMap()
	if err != nil {
		return nil, fmt.Errorf("method RunSelect: error running query: %s", err)
	}

	return rows, nil
}

func RunQuery(queryString string, parameters []interface{}) error {
	query := GetScylla().Query(queryString, nil).Bind(parameters...)

	if err := query.ExecRelease(); err != nil {
		return err
	}

	return nil
}
