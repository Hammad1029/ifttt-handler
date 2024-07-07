package scylla

import (
	"fmt"
	"strings"
)

func RunSelect(queryString string, parameters []interface{}, rowCount int) ([]map[string]interface{}, error) {
	if rowCount != 0 {
		queryString = strings.ReplaceAll(queryString, ";", fmt.Sprintf(" LIMIT %d;", rowCount))
	}

	rows, err := GetScylla().Query(queryString, nil).Bind(parameters).Iter().SliceMap()
	if err != nil {
		return nil, err
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
