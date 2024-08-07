package tables

type Index struct {
	Local     bool     `json:"local" mapstructure:"local"`
	IndexName string   `json:"indexName" mapstructure:"indexName"`
	TableName string   `json:"tableName" mapstructure:"tableName"`
	Columns   []string `json:"columns" mapstructure:"columns"`
}
