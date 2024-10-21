package api

import "context"

type PersistentRepository interface {
	GetAllApis(ctx context.Context) (*[]Api, error)
	GetAllCronJobs(ctx context.Context) (*[]Cron, error)
}
