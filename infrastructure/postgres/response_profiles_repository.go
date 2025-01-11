package infrastructure

import responseprofiles "ifttt/handler/domain/response_profiles"

type PostgresResponseProfilesRepository struct {
	*PostgresBaseRepository
}

func NewPostgresResponseProfilesRepository(base *PostgresBaseRepository) *PostgresResponseProfilesRepository {
	return &PostgresResponseProfilesRepository{PostgresBaseRepository: base}
}

func (p *PostgresResponseProfilesRepository) GetAllInternalProfiles() (*[]responseprofiles.Profile, error) {
	var pgProfiles []response_profile
	p.client.
		Preload("Code").Preload("Description").Preload("Data").
		Preload("MappedProfiles", "parent_id IS NOT NULL").
		Preload("MappedProfiles.Code").Preload("MappedProfiles.Description").Preload("MappedProfiles.Data").
		Where("parent_id IS NULL").
		Find(&pgProfiles)

	dProfiles := make([]responseprofiles.Profile, 0, len(pgProfiles))
	for _, p := range pgProfiles {
		if dP, err := p.toDomain(); err != nil {
			return nil, err
		} else {
			dProfiles = append(dProfiles, *dP)
		}
	}
	return &dProfiles, nil
}
