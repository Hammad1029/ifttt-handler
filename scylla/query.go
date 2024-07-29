package scylla

import (
	"fmt"
	"handler/common"

	jsontocql "github.com/Hammad1029/json-to-cql"
	"github.com/mitchellh/mapstructure"
)

func RunSelect(pQuery jsontocql.ParameterizedQuery, parameters []any) ([]common.JsonObject, error) {
	queryString, err := pQuery.ResolveQuery(parameters)
	if err != nil {
		return nil, fmt.Errorf("method RunSelect: error resolving query: %s", err)
	}

	rows, err := GetScylla().Query(queryString, nil).Iter().SliceMap()
	if err != nil {
		return nil, fmt.Errorf("method RunSelect: error running query: %s", err)
	}

	var results []common.JsonObject
	if err := mapstructure.Decode(rows, &results); err != nil {
		return nil, fmt.Errorf("method RunSelect: could not conver results to []common.JsonObject: %s", err)
	}

	return results, nil
}

func RunQuery(queryString string, parameters []any) error {
	query := GetScylla().Query(queryString, nil).Bind(parameters...)

	if err := query.ExecRelease(); err != nil {
		return err
	}

	return nil
}
