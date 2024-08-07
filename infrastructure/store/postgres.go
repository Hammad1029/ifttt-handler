package infrastructure

import (
	"fmt"
	"handler/common"

	"github.com/mitchellh/mapstructure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStore struct {
	store  *gorm.DB
	config postgresConfig
}

type postgresConfig struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     string `json:"port" mapstructure:"port"`
	Database string `json:"database" mapstructure:"database"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

func (p *PostgresStore) init(config common.JsonObject) error {
	if err := mapstructure.Decode(config, &p.config); err != nil {
		return fmt.Errorf("method: *PostgresStore.Init: could not decode scylla configuration from env: %s", err)
	}
	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Karachi",
		p.config.Host, p.config.Username, p.config.Password, p.config.Database, p.config.Port,
	)
	if db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{}); err != nil {
		return err
	} else {
		p.store = db
	}
	return nil
}
