package infrastructure

import (
	"fmt"
	"ifttt/handler/application/config"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/audit_log"
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
	Store              configStorer
	APIPersistentRepo  api.APIPersistentRepository
	CronPersistentRepo api.CronPersistentRepository
	APIAuditLogRepo    audit_log.ApiAuditLogRepository
	CronAuditLogRepo   audit_log.CronAuditLogRepository
}

type CacheStore struct {
	Store         cacheStorer
	APICacheRepo  api.APICacheRepository
	CronCacheRepo api.CronCacheRepository
}

type DataStore struct {
	Store        dataStorer
	RawQueryRepo resolvable.RawQueryRepository
	DumpRepo     resolvable.DbDumpRepository
}

type AppCacheStore struct {
	Store        appCacheStorer
	AppCacheRepo resolvable.AppCacheRepository
}

func NewConfigStore() (*ConfigStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("configStore")
	if store, err := configStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewDataStore() (*DataStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("dataStore")
	if store, err := dataStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewCacheStore() (*CacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("cacheStore")
	if store, err := cacheStoreFactory(connectionSettings); err != nil {
		return nil, err
	} else {
		return store, nil
	}
}

func NewAppCacheStore() (*AppCacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("cacheStore")
	if store, err := appCacheStoreFactory(connectionSettings); err != nil {
		return nil, err
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
	case postgresDb:
		storer = &postgresStore{}
	// case scyllaDb:
	// 	storer = &scyllaStore{}
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
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("db name not found in env")
	}

	switch strings.ToLower(fmt.Sprint(dbName)) {
	// case scyllaDb:
	// 	storer = &scyllaStore{}
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
	dbName, ok := connectionSettings["db"]
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
		return nil, err
	}

	return storer.createAppCacheStore(), nil
}
