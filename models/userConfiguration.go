package models

import (
	"context"
	"handler/config"
	"handler/scylla"
	"time"

	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type UserConfigurationResolvable struct {
	IsActive          bool      `json:"isActive" mapstructure:"isActive" cql:"is_active"`
	ConfigurationJSON string    `json:"configurationJSON" mapstructure:"configurationJSON" cql:"configuration_json"`
	CreatedAt         time.Time `json:"createdAt" mapstructure:"createdAt" cql:"created_at"`
}

var UserConfigurationMetadata = table.Metadata{
	Name:    "configurations",
	Columns: []string{"is_active", "configuration_json", "created_at"},
	PartKey: []string{"is_active"},
	SortKey: []string{},
}

func (u *UserConfigurationResolvable) Resolve(ctx context.Context) (interface{}, error) {
	return config.GetUserConfig().AllSettings(), nil
}

func (u *UserConfigurationResolvable) ReadUserConfig() error {
	stmt, names := qb.Select("configuration").ToCql()
	q := scylla.GetScylla().Query(stmt, names).BindStruct(u)
	if err := q.SelectRelease(&u); err != nil {
		return err
	}

	return config.SetUserConfig(u.ConfigurationJSON)
}
