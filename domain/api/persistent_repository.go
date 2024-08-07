package api

import "context"

type PersistentRepository interface {
	GetAllApis(ctx context.Context) ([]Api, error)
}
