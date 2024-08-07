package store

import (
	"fmt"
	"handler/common"
	"handler/config"
	"strings"
)

type dbStorer interface {
	init(config common.JsonObject) error
	RawSelect(queryString string, parameters []any) ([]common.JsonObject, error)
	RawQuery(queryString string, parameters []any) ([]common.JsonObject, error)
}

type configStorer interface {
	dbStorer
	GetUserConfiguration(data interface{}) error
}

type dataStorer interface {
	dbStorer
}

type cacheStorer interface {
	init(config common.JsonObject) error
}

var configStore *configStorer
var dataStore *dataStorer
var cacheStore *cacheStorer

func InitConfigStore() error {
	connectionSettings := config.GetConfig().GetStringMap("configStore")
	if store, err := configStoreFactory(connectionSettings); err != nil {
		return fmt.Errorf("method InitConfigStore: could not create store: %s", err)
	} else {
		configStore = store
	}
	return nil
}

func InitDataStore() error {
	connectionSettings := config.GetConfig().GetStringMap("dataStore")
	if store, err := dataStoreFactory(connectionSettings); err != nil {
		return fmt.Errorf("method InitDataStore: could not create store: %s", err)
	} else {
		dataStore = store
	}
	return nil
}

func InitCacheStore() error {
	connectionSettings := config.GetConfig().GetStringMap("cacheStore")
	if store, err := cacheStoreFactory(connectionSettings); err != nil {
		return fmt.Errorf("method InitCacheStore: could not create store: %s", err)
	} else {
		cacheStore = store
	}
	return nil
}

func GetConfigStore() *configStorer {
	return configStore
}

func GetDataStore() *dataStorer {
	return dataStore
}

func GetCacheStore() *cacheStorer {
	return cacheStore
}

func configStoreFactory(connectionSettings common.JsonObject) (*configStorer, error) {
	var store configStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method configStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "scylla":
		store = &ScyllaStore{}
	default:
		return nil, fmt.Errorf("method configStoreFactory: db not found %s", dbName)
	}
	if err := store.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method configStoreFactory: could not init config store: %s", err)
	}
	return &store, nil
}

func dataStoreFactory(connectionSettings common.JsonObject) (*dataStorer, error) {
	var store dataStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method dataStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "scylla":
		store = &ScyllaStore{}
	case "postgres":
		store = &PostgresStore{}
	default:
		return nil, fmt.Errorf("method dataStoreFactory: db not found %s", dbName)
	}
	if err := store.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method dataStoreFactory: could not init config store: %s", err)
	}
	return &store, nil
}

func cacheStoreFactory(connectionSettings common.JsonObject) (*cacheStorer, error) {
	var store cacheStorer
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method cacheStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "redis":
		store = &RedisStore{}
	default:
		return nil, fmt.Errorf("method cacheStoreFactory: db not found %s", dbName)
	}
	if err := store.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method cacheStoreFactory: could not init config store: %s", err)
	}
	return &store, nil
}
