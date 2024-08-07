package adapters

import (
	"github.com/scylladb/gocqlx/v2/table"
)

type scyllaTables struct {
	InternalName   string                 `cql:"internal_name"`
	Name           string                 `cql:"name"`
	Description    string                 `cql:"description"`
	PartitionKeys  []string               `cql:"partition_keys"`
	ClusteringKeys []string               `cql:"clustering_keys"`
	AllColumns     []string               `cql:"all_columns"`
	Mappings       map[string]string      `cql:"mappings"`
	Indexes        map[string]scyllaIndex `cql:"indexes"`
}

var scyllaTablesMetadata = table.Metadata{
	Name:    "Tables",
	Columns: []string{"internal_name", "name", "description", "partition_keys", "clustering_keys", "all_columns", "mappings"},
	PartKey: []string{"internal_name"},
	SortKey: []string{"name", "description"},
}

type scyllaIndex struct {
	Local     bool     `cql:"local"`
	IndexName string   `cql:"index_name"`
	TableName string   `cql:"table_name"`
	Columns   []string `cql:"columns"`
}

type ScyllaTablesRepository struct {
	ScyllaBaseRepository
}

func NewScyllaTablesRepository(base ScyllaBaseRepository) *ScyllaTablesRepository {
	return &ScyllaTablesRepository{ScyllaBaseRepository: base}
}
