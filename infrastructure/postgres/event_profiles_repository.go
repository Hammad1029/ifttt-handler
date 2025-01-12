package infrastructure

import (
	eventprofiles "ifttt/handler/domain/event_profiles"

	"gorm.io/gorm"
)

type PostgresEventProfilesRepository struct {
	*PostgresBaseRepository
}

func NewPostgresEventProfilesRepository(base *PostgresBaseRepository) *PostgresEventProfilesRepository {
	return &PostgresEventProfilesRepository{PostgresBaseRepository: base}
}

func (p *PostgresEventProfilesRepository) GetAllInternalProfiles() (*[]eventprofiles.Profile, error) {
	var pgProfiles []event_profile
	if err := p.client.
		Preload("MappedProfiles", "parent_id IS NOT NULL").
		Where("parent_id IS NULL").Find(&pgProfiles).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	dProfiles := make([]eventprofiles.Profile, 0, len(pgProfiles))
	for _, p := range pgProfiles {
		if dP, err := p.toDomain(); err != nil {
			return nil, err
		} else {
			dProfiles = append(dProfiles, *dP)
		}
	}
	return &dProfiles, nil
}
