package infrastructure

import (
	"fmt"
	"handler/common"
	"handler/domain/configuration"
	"time"

	"github.com/scylladb/gocqlx/v2/qb"
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

func NewScyllaUserConfigurationRepository(base ScyllaBaseRepository) *ScyllaRawQueryRepository {
	return &ScyllaRawQueryRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaRawQueryRepository) GetUserConfigFromDb() (configuration.UserConfiguration, error) {
	var config configuration.UserConfiguration
	stmt, names := qb.Select("configuration").ToCql()
	q := s.session.Query(stmt, names).BindMap(common.JsonObject{"is_active": true})
	if err := q.SelectRelease(&config); err != nil {
		return config,
			fmt.Errorf(
				"method *ScyllaUserConfigurationRepository.GetUserConfigFromDb: error in getting configuration from scylla: %s",
				err,
			)
	}
	return config, nil
}
