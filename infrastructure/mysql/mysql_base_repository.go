package infrastructure

import (
	"database/sql"
)

type MySqlBaseRepository struct {
	client *sql.DB
}

func NewMySqlBaseRepository(client *sql.DB) *MySqlBaseRepository {
	if client == nil {
		panic("missing mysql client")
	}
	return &MySqlBaseRepository{client: client}
}
