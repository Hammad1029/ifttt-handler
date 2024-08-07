package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/domain/api"

	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type scyllaApi struct {
	ApiGroup       string   `cql:"api_group"`
	ApiName        string   `cql:"api_name"`
	ApiDescription string   `cql:"api_description"`
	ApiPath        string   `cql:"api_path"`
	ApiRequest     string   `cql:"api_request"`
	StartRules     []string `cql:"start_rules"`
	Rules          string   `cql:"rules"`
	Queries        string   `cql:"queries"`
}

var scyllaApisMetadata = table.Metadata{
	Name:    "Apis",
	Columns: []string{"api_group", "api_name", "api_description", "api_path", "api_request", "start_rules", "rules", "queries"},
	PartKey: []string{"api_group"},
	SortKey: []string{"api_name", "api_description"},
}

type ScyllaApiPersistentRepository struct {
	ScyllaBaseRepository
}

func NewScyllaApiRepository(base ScyllaBaseRepository) *ScyllaApiPersistentRepository {
	return &ScyllaApiPersistentRepository{ScyllaBaseRepository: base}
}

func (s ScyllaApiPersistentRepository) GetAllApis(ctx context.Context) ([]api.Api, error) {
	var scyllaApis []scyllaApi
	var apis []api.Api

	stmt, names := qb.Select("apis").ToCql()
	q := s.session.Query(stmt, names)
	if err := q.SelectRelease(&scyllaApis); err != nil {
		return nil, fmt.Errorf("method ScyllaApiRepository.GetAllApis: could not get apis: %s", err)
	}

	for _, v := range scyllaApis {
		deserializedApi, err := v.deserialize()
		if err != nil {
			return nil, fmt.Errorf("method ScyllaApiRepository.GetAllApis: failed to deserialize apis: %s", err)
		}
		apis = append(apis, deserializedApi)
	}

	return apis, nil
}

func (a *scyllaApi) deserialize() (api.Api, error) {
	deserializedApi := api.Api{}
	deserializedApi.ApiDescription = a.ApiDescription
	deserializedApi.ApiGroup = a.ApiGroup
	deserializedApi.ApiName = a.ApiName
	deserializedApi.ApiPath = a.ApiPath
	deserializedApi.StartRules = a.StartRules

	err := json.Unmarshal([]byte(a.ApiRequest), &deserializedApi.ApiRequest)
	if err != nil {
		return deserializedApi, fmt.Errorf("method ScyllaApi.deserialize: %s", err)
	}
	err = json.Unmarshal([]byte(a.Rules), &deserializedApi.Rules)
	if err != nil {
		return deserializedApi, fmt.Errorf("method ScyllaApi.deserialize: %s", err)
	}
	err = json.Unmarshal([]byte(a.Queries), &deserializedApi.Queries)
	if err != nil {
		return deserializedApi, fmt.Errorf("method ScyllaApi.deserialize: %s", err)
	}

	return deserializedApi, nil
}
