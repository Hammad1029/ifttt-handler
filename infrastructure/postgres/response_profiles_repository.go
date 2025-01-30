package infrastructure

import (
	"ifttt/handler/domain/configuration"

	"gorm.io/gorm"
)

type PostgresResponseProfilesRepository struct {
	*PostgresBaseRepository
}

func NewPostgresResponseProfilesRepository(base *PostgresBaseRepository) *PostgresResponseProfilesRepository {
	return &PostgresResponseProfilesRepository{PostgresBaseRepository: base}
}

func (r *PostgresResponseProfilesRepository) GetAllProfiles() (*[]configuration.ResponseProfile, error) {
	var pgProfiles []response_profile
	if err := r.client.Find(&pgProfiles).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	dProfiles := make([]configuration.ResponseProfile, 0, len(pgProfiles))
	for _, p := range pgProfiles {
		if dP, err := p.toDomain(); err != nil {
			return nil, err
		} else {
			dProfiles = append(dProfiles, *dP)
		}
	}
	return &dProfiles, nil
}
