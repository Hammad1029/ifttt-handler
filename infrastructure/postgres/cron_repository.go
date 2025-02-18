package infrastructure

import (
	"context"
	"ifttt/handler/domain/api"
)

type PostgresCronRepository struct {
	*PostgresBaseRepository
}

func NewPostgresCronRepository(base *PostgresBaseRepository) *PostgresCronRepository {
	return &PostgresCronRepository{PostgresBaseRepository: base}
}

func (p *PostgresCronRepository) GetAllCronJobs(ctx context.Context) (*[]api.Cron, error) {
	var domainCrons []api.Cron
	var postgresCrons []crons
	if err := p.client.
		Preload("API").Preload("API.Triggers").Preload("API.Triggers.Rules").
		Find(&postgresCrons).Error; err != nil {
		return nil, err
	}

	for _, currPgCron := range postgresCrons {
		if dCron, err := currPgCron.toDomain(); err != nil {
			return nil, err
		} else {
			domainCrons = append(domainCrons, *dCron)
		}
	}

	return &domainCrons, nil
}
