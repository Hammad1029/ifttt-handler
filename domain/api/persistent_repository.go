package api

import "context"

type APIPersistentRepository interface {
	GetAllApis(ctx context.Context) (*[]Api, error)
}

type CronPersistentRepository interface {
	GetAllCronJobs(ctx context.Context) (*[]Cron, error)
}
