package infrastructure

import (
	"ifttt/handler/domain/orm_schema"

	"gorm.io/gorm"
)

type PostgresOrmRepository struct {
	*PostgresBaseRepository
}

func NewPostgresOrmRepository(base *PostgresBaseRepository) *PostgresOrmRepository {
	return &PostgresOrmRepository{PostgresBaseRepository: base}
}

func (o *PostgresOrmRepository) GetAllModels() (*[]orm_schema.Model, error) {
	var pgModels []orm_model
	if err := o.client.
		Preload("Projections").Preload("OwningAssociations").Preload("ReferencedAssociations").
		Preload("OwningAssociations.ReferencesModel").Preload("ReferencedAssociations.OwningModel").
		Find(&pgModels).Error; err == gorm.ErrRecordNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	dModels := []orm_schema.Model{}
	for _, a := range pgModels {
		if dModel, err := a.toDomain(); err != nil {
			return nil, err
		} else {
			dModels = append(dModels, *dModel)
		}
	}
	return &dModels, nil
}

func (o *PostgresOrmRepository) GetAllAssociations() (*[]orm_schema.ModelAssociation, error) {
	var pgAssociation []orm_association
	if err := o.client.Find(&pgAssociation).Error; err == gorm.ErrRecordNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	dAssociations := []orm_schema.ModelAssociation{}
	for _, a := range pgAssociation {
		if dAssociation, err := a.toDomain(); err != nil {
			return nil, err
		} else {
			dAssociations = append(dAssociations, *dAssociation)
		}
	}
	return &dAssociations, nil
}
