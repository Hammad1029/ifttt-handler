package infrastructure

import (
	"context"
	"database/sql"
)

type MySqlRawQueryRepository struct {
	*MySqlBaseRepository
}

func NewMySqlRawQueryRepository(base *MySqlBaseRepository) *MySqlRawQueryRepository {
	return &MySqlRawQueryRepository{MySqlBaseRepository: base}
}

func (m *MySqlRawQueryRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return m.client.BeginTx(ctx, &sql.TxOptions{})
}

func (m *MySqlRawQueryRepository) Scan(tx *sql.Tx, queryString string, parameters []any, ctx context.Context) (*[]map[string]any, int, error) {
	rows, err := tx.QueryContext(ctx, queryString, parameters...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	if mappedRows, err := m.scan(rows); err != nil {
		return nil, 0, err
	} else {
		return mappedRows, len(*mappedRows), nil
	}
}

func (m *MySqlRawQueryRepository) Exec(tx *sql.Tx, queryString string, parameters []any, ctx context.Context) (int, error) {
	if results, err := tx.ExecContext(ctx, queryString, parameters...); err != nil {
		return 0, err
	} else if affected, err := results.RowsAffected(); err != nil {
		return 0, err
	} else {
		return int(affected), nil
	}
}

func (m *MySqlRawQueryRepository) scan(rows *sql.Rows) (*[]map[string]any, error) {
	mappedRows := []map[string]any{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	scanCols := make([]any, len(columns))
	for idx := range scanCols {
		var a any
		scanCols[idx] = &a
	}

	strDescision := make(map[int]bool, len(columns))
	if rows.Next() {
		if err := rows.Scan(scanCols...); err != nil {
			return nil, err
		}

		m := make(map[string]any, len(scanCols))
		for idx, v := range scanCols {
			key := columns[idx]
			retrieved := v.(*any)
			if *retrieved == nil {
				strDescision[idx] = true
				m[key] = nil
			} else if v, ok := (*retrieved).([]byte); ok {
				strDescision[idx] = true
				m[key] = string(v)
			} else {
				m[key] = *retrieved
			}
		}
		mappedRows = append(mappedRows, m)
	}

	for rows.Next() {
		scanCols := make([]any, len(columns))
		for idx := range scanCols {
			var a any
			scanCols[idx] = &a
		}
		if err := rows.Scan(scanCols...); err != nil {
			return nil, err
		}

		m := make(map[string]any, len(columns))
		for idx, v := range scanCols {
			key := columns[idx]
			retrieved := v.(*any)
			if strDescision[idx] {
				if strcast, ok := (*retrieved).([]byte); ok {
					m[key] = string(strcast)
				} else {
					m[key] = *retrieved
				}
			} else {
				m[key] = *retrieved
			}
		}
		mappedRows = append(mappedRows, m)
	}
	return &mappedRows, nil
}
