package infrastructure

import (
	"fmt"
	"strings"
)

type MySqlDbDumpRepository struct {
	*MySqlBaseRepository
}

func NewMySqlDbDumpRepository(base *MySqlBaseRepository) *MySqlDbDumpRepository {
	return &MySqlDbDumpRepository{MySqlBaseRepository: base}
}

func (m *MySqlDbDumpRepository) InsertDump(dump map[string]any, table string) error {
	if len(dump) == 0 {
		return fmt.Errorf("empty data dump")
	}

	columns := make([]string, 0, len(dump))
	placeholders := make([]string, 0, len(dump))
	values := make([]any, 0, len(dump))

	for col, val := range dump {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	if err := m.client.Exec(sql, values...).Error; err != nil {
		return err
	}

	return nil
}
