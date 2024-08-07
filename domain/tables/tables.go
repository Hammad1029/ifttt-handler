package tables

type Tables struct {
	InternalName   string            `json:"internalName" mapstructure:"internalName"`
	Name           string            `json:"name" mapstructure:"name"`
	Description    string            `json:"description" mapstructure:"description"`
	PartitionKeys  []string          `json:"partitionKeys" mapstructure:"partitionKeys"`
	ClusteringKeys []string          `json:"clusteringKeys" mapstructure:"clusteringKeys"`
	AllColumns     []string          `json:"allColumns" mapstructure:"allColumns"`
	Mappings       map[string]string `json:"mappings" mapstructure:"mappings"`
	Indexes        map[string]Index  `json:"indexes" mapstructure:"indexes"`
}
