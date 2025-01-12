package infrastructure

import (
	"fmt"
	postgresInfra "ifttt/handler/infrastructure/postgres"
	"time"

	"github.com/mitchellh/mapstructure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const postgresDb = "postgres"

type postgresStore struct {
	store  *gorm.DB
	config postgresConfig
}

type postgresConfig struct {
	Host     string       `json:"host" mapstructure:"host"`
	Port     string       `json:"port" mapstructure:"port"`
	Database string       `json:"database" mapstructure:"database"`
	Username string       `json:"username" mapstructure:"username"`
	Password string       `json:"password" mapstructure:"password"`
	Pool     postgresPool `json:"pool" mapstructure:"pool"`
}

type postgresPool struct {
	MaxOpenConns int `json:"maxOpenConns" mapstructure:"maxOpenConns"`
	MaxIdleConns int `json:"maxIdleConns" mapstructure:"maxIdleConns"`
	MaxLifeTime  int `json:"maxLifeTime" mapstructure:"maxLifeTime"`
}

func (p *postgresStore) init(config map[string]any) error {
	if err := mapstructure.Decode(config, &p.config); err != nil {
		return fmt.Errorf("method: *PostgresStore.Init: could not decode configuration from env: %s", err)
	}
	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Karachi",
		p.config.Host, p.config.Username, p.config.Password, p.config.Database, p.config.Port,
	)
	if db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}); err != nil {
		return err
	} else {
		p.store = db
		if sqlDb, err := db.DB(); err != nil {
			return err
		} else {
			sqlDb.SetMaxOpenConns(p.config.Pool.MaxOpenConns)
			sqlDb.SetMaxIdleConns(p.config.Pool.MaxIdleConns)
			sqlDb.SetConnMaxLifetime(time.Duration(p.config.Pool.MaxLifeTime) * time.Second)
		}
	}
	return nil
}

func (p *postgresStore) createDataStore() *DataStore {
	postgresBase := postgresInfra.NewPostgresBaseRepository(p.store, false)
	return &DataStore{
		Store:        p,
		RawQueryRepo: postgresInfra.NewPostgresRawQueryRepository(postgresBase),
	}
}

func (p *postgresStore) createConfigStore() *ConfigStore {
	postgresBase := postgresInfra.NewPostgresBaseRepository(p.store, true)
	return &ConfigStore{
		Store:            p,
		APIRepo:          postgresInfra.NewPostgresApiRepository(postgresBase),
		CronRepo:         postgresInfra.NewPostgresCronRepository(postgresBase),
		OrmRepo:          postgresInfra.NewPostgresOrmRepository(postgresBase),
		EventProfileRepo: postgresInfra.NewPostgresEventProfilesRepository(postgresBase),
	}
}
