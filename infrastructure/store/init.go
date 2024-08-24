package infrastructure

import (
	"fmt"
	"handler/application/config"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/configuration"
	"handler/domain/resolvable"
	"handler/domain/tables"
	"strings"
)

type dbStorer interface {
	init(config map[string]any) error
}

type configStorer interface {
	dbStorer
	createConfigStore() *ConfigStore
}

type dataStorer interface {
	dbStorer
	createDataStore() *DataStore
}

type cacheStorer interface {
	dbStorer
	createCacheStore() *CacheStore
}

type ConfigStore struct {
	Store             configStorer
	APIPersistentRepo api.PersistentRepository
	AuditLogRepo      audit_log.Repository
	TablesRepo        tables.Repository
	ConfigRepo        configuration.Repository
}

type DataStore struct {
	Store        dataStorer
	RawQueryRepo resolvable.RawQueryRepository
}

type CacheStore struct {
	Store        cacheStorer
	APICacheRepo api.CacheRepository
}

func NewConfigStore() (*ConfigStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("configStore")
	if store, err := configStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitConfigStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func NewDataStore() (*DataStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("dataStore")
	if store, err := dataStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitDataStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func NewCacheStore() (*CacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("cacheStore")
	if store, err := cacheStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitCacheStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func configStoreFactory(connectionSettings map[string]any) (*ConfigStore, error) {
	var storer configStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method configStoreFactory: db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case scyllaDb:
		storer = &scyllaStore{}
	default:
		return nil, fmt.Errorf("method configStoreFactory: db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method configStoreFactory: could not init config store: %s", err)
	}

	return storer.createConfigStore(), nil
}

func dataStoreFactory(connectionSettings map[string]any) (*DataStore, error) {
	var storer dataStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method dataStoreFactory: db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case scyllaDb:
		storer = &scyllaStore{}
	case postgresDb:
		storer = &postgresStore{}
	case mysqlDb:
		storer = &mysqlStore{}
	default:
		return nil, fmt.Errorf("method dataStoreFactory: db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method dataStoreFactory: could not init data store: %s", err)
	}
	return storer.createDataStore(), nil
}

func cacheStoreFactory(connectionSettings map[string]any) (*CacheStore, error) {
	var storer cacheStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method cacheStoreFactory: db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case redisCache:
		storer = &RedisStore{}
	default:
		return nil, fmt.Errorf("method cacheStoreFactory: db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method configStoreFactory: could not init cache store: %s", err)
	}

	return storer.createCacheStore(), nil
}
