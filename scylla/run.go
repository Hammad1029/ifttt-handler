package scylla

import (
	"fmt"

	"github.com/scylladb/gocqlx"
)

func RunSelect(queryString string, parameters []string, rowCount int) ([]map[string]interface{}, error) {
	if rowCount != 0 {
		queryString += fmt.Sprintf(" LIMIT %d", rowCount)
	}

	query := GetScylla().Query(queryString, parameters)
	defer query.Release()

	var res []map[string]interface{}
	if err := gocqlx.Select(&res, query.Query); err != nil {
		return nil, err
	}
	return res, nil
}

func RunQuery(queryString string, parameters []string) error {
	query := GetScylla().Query(queryString, parameters)
	defer query.Release()

	if err := query.ExecRelease(); err != nil {
		return err
	}

	return nil
}
