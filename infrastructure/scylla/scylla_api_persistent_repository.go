package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/domain/api"

	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaApi struct {
	ApiGroup       string   `cql:"api_group"`
	ApiName        string   `cql:"api_name"`
	ApiDescription string   `cql:"api_description"`
	ApiPath        string   `cql:"api_path"`
	ApiRequest     string   `cql:"api_request"`
	Rules          string   `cql:"rules"`
	StartRules     []string `cql:"start_rules"`
}

var scyllaApisMetadata = table.Metadata{
	Name:    "apis",
	Columns: []string{"api_group", "api_name", "api_description", "api_path", "api_request", "rules", "start_rules"},
	PartKey: []string{"api_group"},
	SortKey: []string{"api_name", "api_description"},
}

var scyllaApisTable *table.Table

type ScyllaApiPersistentRepository struct {
	ScyllaBaseRepository
}

func NewScyllaApiPersistentRepository(base ScyllaBaseRepository) *ScyllaApiPersistentRepository {
	return &ScyllaApiPersistentRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaApiPersistentRepository) getTable() *table.Table {
	if scyllaApisTable == nil {
		scyllaApisTable = table.New(scyllaApisMetadata)
	}
	return scyllaApisTable
}

func (s *ScyllaApiPersistentRepository) GetAllApis(ctx context.Context) (*[]api.Api, error) {
	var scyllaApis []scyllaApi
	apis := &([]api.Api{})

	apisTable := s.getTable()
	stmt, names := apisTable.SelectAll()
	if err := s.session.Query(stmt, names).SelectRelease(&scyllaApis); err != nil {
		return nil, fmt.Errorf("method *ScyllaApiPersistentRepository.GetAllApis: could not get apis: %s", err)
	}

	for _, v := range scyllaApis {
		deserializedApi, err := v.deserialize()
		if err != nil {
			return nil, fmt.Errorf("method *ScyllaApiPersistentRepository.GetAllApis: failed to deserialize apis: %s", err)
		}
		*apis = append(*apis, *deserializedApi)
	}

	return apis, nil
}

func (a *scyllaApi) deserialize() (*api.Api, error) {
	deserializedApi := api.Api{}
	deserializedApi.ApiDescription = a.ApiDescription
	deserializedApi.ApiGroup = a.ApiGroup
	deserializedApi.ApiName = a.ApiName
	deserializedApi.ApiPath = a.ApiPath
	deserializedApi.StartRules = a.StartRules

	err := json.Unmarshal([]byte(a.ApiRequest), &deserializedApi.ApiRequest)
	if err != nil {
		return nil, fmt.Errorf("method ScyllaApi.deserialize: could not deserialize api request: %s", err)
	}
	err = json.Unmarshal([]byte(a.Rules), &deserializedApi.Rules)
	if err != nil {
		return nil, fmt.Errorf("method ScyllaApi.deserialize: could not deserialize rules: %s", err)
	}
	return &deserializedApi, nil
}
