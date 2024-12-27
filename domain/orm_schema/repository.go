package orm_schema

import (
	"context"
)

type PersistentRepository interface {
	GetAllModels() (*[]Model, error)
	GetAllAssociations() (*[]ModelAssociation, error)
}

type CacheRepository interface {
	GetModel(name string, ctx context.Context) (*Model, error)
	SetModels(schemas *map[string]Model, ctx context.Context) error

	GetAssociation(name string, ctx context.Context) (*ModelAssociation, error)
	SetAssociations(associations *map[string]ModelAssociation, ctx context.Context) error
}
