package api

import "context"

type CacheRepository interface {
	StoreApis(apis *[]ApiSerialized, ctx context.Context) error
	GetAllApis(ctx context.Context) (*[]ApiSerialized, error)
	GetApiByPath(path string, ctx context.Context) (*ApiSerialized, error)
}
