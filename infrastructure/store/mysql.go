package infrastructure

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	mysqlInfra "ifttt/handler/infrastructure/mysql"

	"github.com/mitchellh/mapstructure"
)

const mysqlDb = "mysql"

type mysqlStore struct {
	store  *sql.DB
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
		return fmt.Errorf("method: *mysqlStore.Init: could not decode configuration from env: %s", err)
	}
	connectionString := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		m.config.Username, m.config.Password, m.config.Host, m.config.Port, m.config.Database,
	)
	if db, err := sql.Open(mysqlDb, connectionString); err != nil {
		return err
	} else {
		m.store = db
	}
	return nil
}

func (m *mysqlStore) createDataStore() *DataStore {
	mysqlBase := mysqlInfra.NewMySqlBaseRepository(m.store)
	return &DataStore{
		Store:        m,
		RawQueryRepo: mysqlInfra.NewMySqlRawQueryRepository(mysqlBase),
	}
}
