package infrastructure

import (
	"fmt"
	"handler/common"
	"handler/config"
	"handler/domain/api"
	"handler/domain/audit_log"
	"handler/domain/configuration"
	"handler/domain/resolvable"
	"handler/domain/tables"
	postgresInfra "handler/infrastructure/postgres"
	redisInfra "handler/infrastructure/redis"
	scyllaInfra "handler/infrastructure/scylla"
	"strings"
)

type dbStorer interface {
	init(config common.JsonObject) error
}

type ConfigStorer interface {
	dbStorer
}

type ConfigStore struct {
	Store             ConfigStorer
	APIPersistentRepo api.PersistentRepository
	AuditLogRepo      audit_log.Repository
	TablesRepo        tables.Repository
	ConfigRepo        configuration.Repository
}

type DataStorer interface {
	dbStorer
}

type DataStore struct {
	Store DataStorer
	resolvable.RawQueryRepository
}

type CacheStorer interface {
	init(config common.JsonObject) error
}

type CacheStore struct {
	Store        CacheStorer
	APICacheRepo api.CacheRepository
}

type MainStore struct {
	ConfigStore       ConfigStore
	DataStore         DataStore
	CacheStore        CacheStore
	UserConfiguration configuration.UserConfiguration
}

func NewConfigStore() (*ConfigStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("configStore")
	if store, err := configStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitConfigStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func configStoreFactory(connectionSettings common.JsonObject) (*ConfigStore, error) {
	var store ConfigStore
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method configStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "scylla":
		{
			scyllaStore := ScyllaStore{}
			if err := scyllaStore.init(connectionSettings); err != nil {
				return nil, fmt.Errorf("method configStoreFactory: could not init config store: %s", err)
			}
			scyllaBase := scyllaInfra.NewScyllaBaseRepository(scyllaStore.session, scyllaStore.cluster)
			store = ConfigStore{
				Store:             &scyllaStore,
				APIPersistentRepo: scyllaInfra.NewScyllaApiRepository(*scyllaBase),
				AuditLogRepo:      scyllaInfra.NewScyllaAuditLogRepository(*scyllaBase),
				TablesRepo:        scyllaInfra.NewScyllaTablesRepository(*scyllaBase),
				ConfigRepo:        scyllaInfra.NewScyllaUserConfigurationRepository(*scyllaBase),
			}
		}
	default:
		return nil, fmt.Errorf("method configStoreFactory: db not found %s", dbName)
	}
	return &store, nil
}

func NewDataStore() (*DataStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("dataStore")
	if store, err := dataStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitDataStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func dataStoreFactory(connectionSettings common.JsonObject) (*DataStore, error) {
	var store DataStore
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method dataStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "scylla":
		{
			scyllaStore := ScyllaStore{}
			if err := scyllaStore.init(connectionSettings); err != nil {
				return nil, fmt.Errorf("method dataStoreFactory: could not init data store: %s", err)
			}
			scyllaBase := scyllaInfra.NewScyllaBaseRepository(scyllaStore.session, scyllaStore.cluster)
			store = DataStore{
				Store:              &scyllaStore,
				RawQueryRepository: scyllaInfra.NewScyllaRawQueryRepository(*scyllaBase),
			}
		}
	case "postgres":
		{
			postgresStore := PostgresStore{}
			if err := postgresStore.init(connectionSettings); err != nil {
				return nil, fmt.Errorf("method dataStoreFactory: could not init data store: %s", err)
			}
			postgresBase := postgresInfra.NewPostgresBaseRepository(postgresStore.store)
			store = DataStore{
				Store:              &postgresStore,
				RawQueryRepository: postgresInfra.NewPostgresRawQueryRepository(*postgresBase),
			}
		}
	default:
		return nil, fmt.Errorf("method dataStoreFactory: db not found %s", dbName)
	}
	if err := store.Store.init(connectionSettings); err != nil {
		return nil, fmt.Errorf("method dataStoreFactory: could not init config store: %s", err)
	}
	return &store, nil
}

func NewCacheStore() (*CacheStore, error) {
	connectionSettings := config.GetConfig().GetStringMap("cacheStore")
	if store, err := cacheStoreFactory(connectionSettings); err != nil {
		return nil, fmt.Errorf("method InitCacheStore: could not create store: %s", err)
	} else {
		return store, nil
	}
}

func cacheStoreFactory(connectionSettings common.JsonObject) (*CacheStore, error) {
	var store CacheStore
	dbName, ok := connectionSettings["db"]
	if !ok {
		return nil, fmt.Errorf("method cacheStoreFactory: db name not found in env")
	}
	switch strings.ToLower(fmt.Sprint(dbName)) {
	case "redis":
		{
			redisStore := RedisStore{}
			if err := redisStore.init(connectionSettings); err != nil {
				return nil, fmt.Errorf("method configStoreFactory: could not init config store: %s", err)
			}
			redisBase := redisInfra.NewRedisBaseRepository(redisStore.client)
			store = CacheStore{
				Store:        &redisStore,
				APICacheRepo: redisInfra.NewRedisApiCacheRepository(*redisBase),
			}
		}
	default:
		return nil, fmt.Errorf("method cacheStoreFactory: db not found %s", dbName)
	}
	return &store, nil
}
