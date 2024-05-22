package scylla

import (
	"handler/config"
	"handler/utils"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

var cluster *gocql.ClusterConfig
var scyllaSession *gocqlx.Session

func Init() {
	configuration := config.GetConfig()
	createCluster(
		gocql.Quorum,
		config.GetConfigProp("scylla.keyspace"),
		configuration.GetStringSlice("scylla.nodes")...,
	)

	session, err := gocqlx.WrapSession(gocql.NewSession(*cluster))
	scyllaSession = &session
	utils.HandleError(err)

	// Right now perform initial DDL manually

	// create keyspace IF NOT EXISTS
	// err = session.ExecStmt(GetSchemasProp("keyspace"))
	// utils.HandleError(err)

	// // create table for tables IF NOT EXISTS
	// err = session.ExecStmt(GetSchemasProp("tables.tables"))
	// utils.HandleError(err)

	// defer session.Close()
}

func GetScylla() *gocqlx.Session {
	return scyllaSession
}

func createCluster(consistency gocql.Consistency, keyspace string, hosts ...string) {
	config := config.GetConfig()
	retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
		Min:        time.Duration(config.GetInt("scylla.retry.min")) * time.Second,
		Max:        time.Duration(config.GetInt("scylla.retry.max")) * time.Second,
		NumRetries: config.GetInt("scylla.retry.noOfRetries"),
	}
	cluster = gocql.NewCluster(hosts...)
	cluster.Keyspace = keyspace
	cluster.Timeout = time.Duration(config.GetInt("scylla.timeout")) * time.Second
	cluster.RetryPolicy = retryPolicy
	cluster.Consistency = consistency
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
}
