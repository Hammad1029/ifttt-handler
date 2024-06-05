package scylla

import (
	"fmt"
	"handler/utils"
	"strings"
)

func RunSelect(queryString string, parameters []string, rowCount int) ([]map[string]interface{}, error) {
	if rowCount != 0 {
		queryString = strings.ReplaceAll(queryString, ";", fmt.Sprintf(" LIMIT %d;", rowCount))
	}

	queryString = "SELECT int1, text1 FROM table_table1_62566050_185f_11ef_bf29_001 WHERE int1=1 and text2='yooo';"

	query := GetScylla().Query(queryString, nil).Bind(utils.ConvertStringToInterfaceArray(parameters)...)
	var res [][]interface{}
	if err := query.SelectRelease(&res); err != nil {
		return nil, err
	}

	return nil, nil

}

func RunQuery(queryString string, parameters []string) error {
	query := GetScylla().Query(queryString, parameters)
	defer query.Release()

	if err := query.ExecRelease(); err != nil {
		return err
	}

	return nil
}
