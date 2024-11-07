package infrastructure

import (
	"context"
	"ifttt/handler/domain/api"
)

type PostgresApiRepository struct {
	*PostgresBaseRepository
}

func NewPostgresApiRepository(base *PostgresBaseRepository) *PostgresApiRepository {
	return &PostgresApiRepository{PostgresBaseRepository: base}
}

func (p *PostgresApiRepository) GetAllApis(ctx context.Context) (*[]api.Api, error) {
	var domainApis []api.Api
	var postgresApis []apis
	if err := p.client.
		Preload("PreWare").Preload("PreWare.Rules").
		Preload("MainWare").Preload("MainWare.Rules").
		Preload("PostWare").Preload("PostWare.Rules").
		Find(&postgresApis).Error; err != nil {
		return nil, err
	}

	for _, currPgApi := range postgresApis {
		if dApi, err := currPgApi.toDomain(); err != nil {
			return nil, err
		} else {
			domainApis = append(domainApis, *dApi)
		}
	}

	return &domainApis, nil
}
