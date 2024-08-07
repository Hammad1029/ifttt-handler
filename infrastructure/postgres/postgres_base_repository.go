package infrastructure

import (
	"gorm.io/gorm"
)

type PostgresBaseRepository struct {
	client *gorm.DB
}

func NewPostgresBaseRepository(client *gorm.DB) *PostgresBaseRepository {
	if client == nil {
		panic("missing redis client")
	}
	return &PostgresBaseRepository{client: client}
}
