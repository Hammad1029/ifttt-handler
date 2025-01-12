package eventprofiles

import "context"

type PersistentRepository interface {
	GetAllInternalProfiles() (*[]Profile, error)
}

type CacheRepository interface {
	StoreProfiles(profiles *map[string]Profile, ctx context.Context) error
	GetProfileByTrigger(trigger string, ctx context.Context) (*Profile, error)
}
