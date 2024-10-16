package infrastructure

import (
	"context"
	"fmt"
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
		Preload("TriggerFlowRef").Preload("TriggerFlowRef.StartRules").Preload("TriggerFlowRef.AllRules").
		Find(&postgresApis).Error; err != nil {
		return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not get apis from postgres: %s", err)
	}

	for _, currPgApi := range postgresApis {
		if dApi, err := currPgApi.toDomain(); err != nil {
			return nil, fmt.Errorf("method *PostgresApiRepository.GetAllApis: could not convert to domain api: %s", err)
		} else {
			domainApis = append(domainApis, *dApi)
		}
	}

	return &domainApis, nil
}
