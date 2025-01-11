package responseprofiles

import "context"

type PersistentRepository interface {
	GetAllInternalProfiles() (*[]Profile, error)
}

type CacheRepository interface {
	StoreProfiles(profiles *map[string]Profile, ctx context.Context) error
	GetProfileByCode(code string, ctx context.Context) (*Profile, error)
}
