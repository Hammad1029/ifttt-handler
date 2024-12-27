package orm_schema

type Model struct {
	Name                   string             `mapstructure:"name" json:"name"`
	Table                  string             `mapstructure:"table" json:"table"`
	Projections            []Projection       `mapstructure:"projections" json:"projections"`
	PrimaryKey             string             `mapstructure:"primaryKey" json:"primaryKey"`
	OwningAssociations     []ModelAssociation `mapstructure:"owningAssociations" json:"owningAssociations"`
	ReferencedAssociations []ModelAssociation `mapstructure:"referencedAssociations" json:"referencedAssociations"`
}

type Projection struct {
	Column   string `mapstructure:"column" json:"column"`
	As       string `mapstructure:"as" json:"as"`
	DataType string `mapstructure:"dataType" json:"dataType"`
}

type ModelAssociation struct {
	Name                 string `mapstructure:"name" json:"name"`
	Type                 string `mapstructure:"type" json:"type"`
	TableName            string `mapstructure:"tableName" json:"tableName"`
	ColumnName           string `mapstructure:"columnName" json:"columnName"`
	ReferencesTable      string `mapstructure:"referencesTable" json:"referencesTable"`
	ReferencesField      string `mapstructure:"referencesField" json:"referencesField"`
	JoinTable            string `mapstructure:"joinTable" json:"joinTable"`
	JoinTableSourceField string `mapstructure:"joinTableSourceField" json:"joinTableSourceField"`
	JoinTableTargetField string `mapstructure:"joinTableTargetField" json:"joinTableTargetField"`
	OwningModelID        uint   `mapstructure:"owningModelID" json:"owningModelID"`
	ReferencesModelID    uint   `mapstructure:"referencesModelID" json:"referencesModelID"`
	OwningModel          Model  `mapstructure:"owningModel" json:"owningModel"`
	ReferencesModel      Model  `mapstructure:"referencesModel" json:"referencesModel"`
}

type Populate struct {
	Model    string     `mapstructure:"model" json:"model"`
	As       string     `mapstructure:"as" json:"as"`
	Populate []Populate `mapstructure:"populate" json:"populate"`
}
