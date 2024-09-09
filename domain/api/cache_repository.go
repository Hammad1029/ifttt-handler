package api

import "context"

type CacheRepository interface {
	StoreApis(apis *[]Api, ctx context.Context) error
	GetAllApis(ctx context.Context) (*[]Api, error)
	GetApiByPath(path string, ctx context.Context) (*Api, error)
}
