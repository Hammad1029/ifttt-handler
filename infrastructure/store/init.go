package infrastructure

import (
	"fmt"
	"ifttt/handler/application/config"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/orm_schema"
	"ifttt/handler/domain/resolvable"
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

type appCacheStorer interface {
	dbStorer
	createAppCacheStore() *AppCacheStore
}

type ConfigStore struct {
	Store    configStorer
	APIRepo  api.APIPersistentRepository
	CronRepo api.CronPersistentRepository
	OrmRepo  orm_schema.PersistentRepository
}

type CacheStore struct {
	Store    cacheStorer
	APIRepo  api.APICacheRepository
	CronRepo api.CronCacheRepository
	OrmRepo  orm_schema.CacheRepository
}

type DataStore struct {
	Store        dataStorer
	RawQueryRepo resolvable.RawQueryRepository
}

type AppCacheStore struct {
	Store        appCacheStorer
	AppCacheRepo resolvable.AppCacheRepository
}

func NewConfigStore() (*ConfigStore, error) {
	connectionSettings := config.GetConfig().GetStringMap(common.EnvConfig)
	if store, err := configStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewDataStore() (*DataStore, error) {
	connectionSettings := config.GetConfig().GetStringMap(common.EnvData)
	if store, err := dataStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewCacheStore() (*CacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap(common.EnvCache)
	if store, err := cacheStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewAppCacheStore() (*AppCacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap(common.EnvAppCache)
	if store, err := appCacheStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func configStoreFactory(connectionSettings map[string]any) (*ConfigStore, error) {
	var storer configStorer
	dbName, ok := connectionSettings[common.EnvDBName]
	if !ok {
		return nil, fmt.Errorf("method configStoreFactory: db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case postgresDb:
		storer = &postgresStore{}
	default:
		return nil, fmt.Errorf("db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("could not init config store: %s", err)
	}

	return storer.createConfigStore(), nil
}

func dataStoreFactory(connectionSettings map[string]any) (*DataStore, error) {
	var storer dataStorer
	dbName, ok := connectionSettings[common.EnvDBName]
	if !ok {
		return nil, fmt.Errorf("db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case postgresDb:
		storer = &postgresStore{}
	case mysqlDb:
		storer = &mysqlStore{}
	default:
		return nil, fmt.Errorf("db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, err
	}
	return storer.createDataStore(), nil
}

func cacheStoreFactory(connectionSettings map[string]any) (*CacheStore, error) {
	var storer cacheStorer
	dbName, ok := connectionSettings[common.EnvDBName]
	if !ok {
		return nil, fmt.Errorf("db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	case redisCache:
		storer = &RedisStore{}
	default:
		return nil, fmt.Errorf("db not found %s", dbName)
	}

	if err := storer.init(connectionSettings); err != nil {
		return nil, err
	}

	return storer.createCacheStore(), nil
}

func appCacheStoreFactory(connectionSettings map[string]any) (*AppCacheStore, error) {
	var storer appCacheStorer
	dbName, ok := connectionSettings[common.EnvDBName]
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
		return nil, err
	}

	return storer.createAppCacheStore(), nil
}
