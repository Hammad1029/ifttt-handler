package infrastructure

import (
	"fmt"
	"ifttt/handler/domain/orm_schema"
)

type MySqlSchemaRepository struct {
	*MySqlBaseRepository
}

func NewMySqlOrmSchemaRepository(base *MySqlBaseRepository) *MySqlSchemaRepository {
	return &MySqlSchemaRepository{MySqlBaseRepository: base}
}

func (p *MySqlSchemaRepository) GetTableNames() ([]string, error) {
	var names []string
	if err := p.client.Table("information_schema.tables").Where(
		"table_schema = DATABASE()",
	).Pluck("table_name", &names).Error; err != nil {
		return nil,
			fmt.Errorf("method *PostgresSchemaRepository.GetAllTables: could not get table names: %s", err)
	}
	return names, nil
}

func (p *MySqlSchemaRepository) GetAllColumns(tables []string) (*[]orm_schema.Column, error) {
	var columns []orm_schema.Column
	if err := p.client.Table("information_schema.columns").
		Select(
			"TABLE_NAME as TableName, "+
				"ORDINAL_POSITION as OrdinalPosition, "+
				"COLUMN_NAME as ColumnName, "+
				"DATA_TYPE as DataType, "+
				"COLUMN_DEFAULT as ColumnDefault, "+
				"IS_NULLABLE as IsNullable, "+
				"CHARACTER_MAXIMUM_LENGTH as CharacterMaximumLength, "+
				"NUMERIC_PRECISION as NumericPrecision",
		).
		Where("table_name IN ?", tables).
		Order("ordinal_position").
		Scan(&columns).Error; err != nil {
		return nil,
			fmt.Errorf("could not get columns: %s", err)
	}
	return &columns, nil
}

func (p *MySqlSchemaRepository) GetAllConstraints(tables []string) (*[]orm_schema.Constraint, error) {
	var constraints []orm_schema.Constraint
	if err := p.client.Table("information_schema.table_constraints AS tc").
		Select(
			"tc.table_name as TableName, "+
				"tc.constraint_name as ConstraintName, "+
				"tc.constraint_type as ConstraintType, "+
				"kcu.column_name as ColumnName, "+
				"kcu.referenced_table_name as ReferencesTable, "+
				"kcu.referenced_column_name as ReferencesField",
		).
		Joins("LEFT JOIN information_schema.key_column_usage AS kcu ON "+
			"kcu.table_schema = tc.table_schema AND "+
			"kcu.table_name = tc.table_name AND "+
			"kcu.constraint_name = tc.constraint_name").
		Where("tc.table_schema = DATABASE() AND tc.table_name IN ?", tables).
		Scan(&constraints).Error; err != nil {
		return nil,
			fmt.Errorf("method *PostgresSchemaRepository.GetAllColumns: could not get columns: %s", err)
	}
	return &constraints, nil
}
