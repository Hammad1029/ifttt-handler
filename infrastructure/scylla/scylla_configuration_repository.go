package infrastructure

import (
	"fmt"
	"ifttt/handler/domain/configuration"
	"time"

	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaConfiguration struct {
	IsActive          bool      `cql:"is_active"`
	ConfigurationJSON string    `cql:"configuration_json"`
	CreatedAt         time.Time `cql:"created_at"`
}

var configurationMetadata = table.Metadata{
	Name:    "configurations",
	Columns: []string{"is_active", "configuration_json", "created_at"},
	PartKey: []string{"is_active"},
	SortKey: []string{},
}

var scyllaConfigurationTable *table.Table

type ScyllaConfigurationRepository struct {
	ScyllaBaseRepository
}

func NewScyllaConfigurationRepository(base ScyllaBaseRepository) *ScyllaConfigurationRepository {
	return &ScyllaConfigurationRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaConfigurationRepository) getTable() *table.Table {
	if scyllaConfigurationTable == nil {
		scyllaConfigurationTable = table.New(configurationMetadata)
	}
	return scyllaConfigurationTable
}

func (s *ScyllaConfigurationRepository) GetConfigFromDb() (*configuration.Configuration, error) {
	var configs []configuration.Configuration

	configurationTable := s.getTable()
	query := configurationTable.SelectQuery(*s.session).BindStruct(scyllaConfiguration{IsActive: true})
	if err := query.SelectRelease(&configs); err != nil {
		return nil, fmt.Errorf(
			"method *ScyllaConfigurationRepository.GetConfigFromDb: error in getting configuration from scylla: %s",
			err,
		)
	}

	if len(configs) == 0 {
		return nil, nil
	}
	return &(configs[0]), nil
}
