package orm_schema

import (
	"context"
)

type PersistentRepository interface {
	GetTableNames() ([]string, error)
	GetAllColumns(tables []string) (*[]Column, error)
	GetAllConstraints(tables []string) (*[]Constraint, error)
}

type CacheRepository interface {
	GetSchema(tableName string, ctx context.Context) (*Schema, error)
	StoreSchema(schemas *[]Schema, ctx context.Context) error
}
