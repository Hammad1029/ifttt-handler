package infrastructure

import (
	"fmt"
	"strings"
)

type PostgresDbDumpRepository struct {
	*PostgresBaseRepository
}

func NewPostgresDbDumpRepository(base *PostgresBaseRepository) *PostgresDbDumpRepository {
	return &PostgresDbDumpRepository{PostgresBaseRepository: base}
}

func (p *PostgresDbDumpRepository) InsertDump(dump map[string]any, table string) error {
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
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	if err := p.client.Exec(sql, values...).Error; err != nil {
		return err
	}

	return nil
}
