package infrastructure

import (
	"context"
	"fmt"
	"ifttt/handler/domain/api"

	"github.com/mitchellh/mapstructure"
	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaApiSerialized struct {
	Group       string   `cql:"group" mapstructure:"group"`
	Name        string   `cql:"name" mapstructure:"name"`
	Method      string   `cql:"method" mapstructure:"method"`
	Type        string   `cql:"type" mapstructure:"type"`
	Path        string   `cql:"path" mapstructure:"path"`
	Description string   `cql:"description" mapstructure:"description"`
	Request     string   `cql:"request" mapstructure:"request"`
	Dumping     string   `cql:"dumping" mapstructure:"dumping"`
	StartRules  []string `cql:"start_rules" mapstructure:"rules"`
	Rules       string   `cql:"rules" mapstructure:"startRules"`
}

var scyllaApisMetadata = table.Metadata{
	Name:    "apis",
	Columns: []string{"group", "name", "method", "type", "path", "description", "request", "dumping", "start_rules", "rules"},
	PartKey: []string{"group"},
	SortKey: []string{"name"},
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

func (s *ScyllaApiPersistentRepository) GetAllApis(ctx context.Context) (*[]api.ApiSerialized, error) {
	var scyllaApis []scyllaApiSerialized
	serializedApis := &([]api.ApiSerialized{})

	apisTable := s.getTable()
	stmt, names := apisTable.SelectAll()
	if err := s.session.Query(stmt, names).SelectRelease(&scyllaApis); err != nil {
		return nil, fmt.Errorf("method *ScyllaApiPersistentRepository.GetAllApis: could not get apis: %s", err)
	}

	if err := mapstructure.Decode(scyllaApis, &serializedApis); err != nil {
		return nil, fmt.Errorf("method *ScyllaApiPersistentRepository.GetAllApis: failed to decode apis: %s", err)
	}

	return serializedApis, nil
}
