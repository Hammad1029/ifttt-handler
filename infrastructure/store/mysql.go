package infrastructure

import (
	"fmt"
	postgresInfra "handler/infrastructure/postgres"

	"github.com/mitchellh/mapstructure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlStore struct {
	store  *gorm.DB
	config mysqlConfig
}

type mysqlConfig struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     string `json:"port" mapstructure:"port"`
	Database string `json:"database" mapstructure:"database"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

func (m *mysqlStore) init(config map[string]any) error {
	if err := mapstructure.Decode(config, &m.config); err != nil {
		return fmt.Errorf("method: *mysqlStore.Init: could not decode scylla configuration from env: %s", err)
	}
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		m.config.Username, m.config.Password, m.config.Host, m.config.Port, m.config.Database,
	)
	if db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{}); err != nil {
		return err
	} else {
		m.store = db
	}
	return nil
}

func (m *mysqlStore) createDataStore() *DataStore {
	postgresBase := postgresInfra.NewPostgresBaseRepository(m.store)
	return &DataStore{
		Store:        m,
		RawQueryRepo: postgresInfra.NewPostgresRawQueryRepository(*postgresBase),
	}
}
