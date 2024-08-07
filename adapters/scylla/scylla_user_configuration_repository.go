package adapters

import (
	"time"

	"github.com/scylladb/gocqlx/v2/table"
)

type scyllaUserConfiguration struct {
	IsActive          bool      `cql:"is_active"`
	ConfigurationJSON string    `cql:"configuration_json"`
	CreatedAt         time.Time `cql:"created_at"`
}

var UserConfigurationMetadata = table.Metadata{
	Name:    "configurations",
	Columns: []string{"is_active", "configuration_json", "created_at"},
	PartKey: []string{"is_active"},
	SortKey: []string{},
}

type ScyllaUserConfigurationRepository struct {
	ScyllaBaseRepository
}

func NewScyllaUserConfigurationRepository(base ScyllaBaseRepository) *ScyllaUserConfigurationRepository {
	return &ScyllaUserConfigurationRepository{ScyllaBaseRepository: base}
}
