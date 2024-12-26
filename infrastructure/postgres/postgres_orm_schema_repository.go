package infrastructure

import (
	"fmt"
	"ifttt/handler/domain/orm_schema"
)

type PostgresSchemaRepository struct {
	*PostgresBaseRepository
}

func NewPostgresOrmSchemaRepository(base *PostgresBaseRepository) *PostgresSchemaRepository {
	return &PostgresSchemaRepository{PostgresBaseRepository: base}
}

func (p *PostgresSchemaRepository) GetTableNames() ([]string, error) {
	var names []string
	if err := p.client.Table("information_schema.tables").Where(
		"table_type = ? AND table_schema NOT IN ?",
		"BASE TABLE", []string{"pg_catalog", "information_schema"},
	).Pluck("table_name", &names).Error; err != nil {
		return nil,
			fmt.Errorf("method *PostgresSchemaRepository.GetAllTables: could not get table names: %s", err)
	}
	return names, nil
}

func (p *PostgresSchemaRepository) GetAllColumns(tables []string) (*[]orm_schema.Column, error) {
	var columns []orm_schema.Column
	if err := p.client.Table("information_schema.columns").
		Select("table_name,ordinal_position,column_name,data_type,column_default,is_nullable,character_maximum_length,numeric_precision").
		Where("table_name IN ?", tables).
		Order("ordinal_position").
		Scan(&columns).Error; err != nil {
		return nil,
			fmt.Errorf("method *PostgresSchemaRepository.GetAllColumns: could not get columns: %s", err)
	}
	return &columns, nil
}

func (p *PostgresSchemaRepository) GetAllConstraints(tables []string) (*[]orm_schema.Constraint, error) {
	var constraints []orm_schema.Constraint
	if err := p.client.Table("information_schema.table_constraints AS tc").
		Select("tc.constraint_name, tc.constraint_type, tc.table_name, kcu.column_name, ccu.table_name AS references_table, ccu.column_name AS references_field").
		Joins("LEFT JOIN information_schema.key_column_usage AS kcu ON tc.constraint_catalog = kcu.constraint_catalog AND tc.constraint_schema = kcu.constraint_schema AND tc.constraint_name = kcu.constraint_name").
		Joins("LEFT JOIN information_schema.referential_constraints AS rc ON tc.constraint_catalog = rc.constraint_catalog AND tc.constraint_schema = rc.constraint_schema AND tc.constraint_name = rc.constraint_name").
		Joins("LEFT JOIN information_schema.constraint_column_usage AS ccu ON rc.unique_constraint_catalog = ccu.constraint_catalog AND rc.unique_constraint_schema = ccu.constraint_schema AND rc.unique_constraint_name = ccu.constraint_name").
		Where("tc.table_name in ?", tables).
		Scan(&constraints).Error; err != nil {
		return nil,
			fmt.Errorf("method *PostgresSchemaRepository.GetAllColumns: could not get columns: %s", err)
	}
	return &constraints, nil
}
