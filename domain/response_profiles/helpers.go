package responseprofiles

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

func (f *Field) AddToMap(compare any, m *map[string]any) {
	if !f.Disabled {
		if compare != nil && compare != "" {
			(*m)[f.Key] = compare
		} else {
			(*m)[f.Key] = f.Default
		}
	}
}

func transformProfiles(profiles *[]Profile) map[string]Profile {
	if profiles == nil {
		return nil
	}
	transformed := make(map[string]Profile, len(*profiles))
	for _, p := range *profiles {
		transformed[p.MappedCode] = p
	}
	return transformed
}

func GetAndStoreProfiles(
	persistent PersistentRepository, cache CacheRepository, ctx context.Context) error {
	profiles, err := persistent.GetAllInternalProfiles()
	if err != nil {
		return err
	}
	transformed := transformProfiles(profiles)
	if transformed == nil {
		return fmt.Errorf("no response profiles found")
	} else {
		for _, rc := range common.ResponseCodes {
			if profile, ok := transformed[rc]; !ok {
				return fmt.Errorf("response profile not found for code: %s", rc)
			} else if profile.Code.Key != common.InternalResponseCode {
				return fmt.Errorf("invalid code key for response profile %s: %s", rc, profile.Code.Key)
			} else if profile.Description.Key != common.InternalResponseDescription {
				return fmt.Errorf("invalid description key for response profile %s: %s", rc, profile.Description.Key)
			} else if profile.Data.Key != common.InternalResponseData {
				return fmt.Errorf("invalid data key for response profile %s: %s", rc, profile.Data.Key)
			} else if profile.Errors.Key != common.InternalResponseErrors {
				return fmt.Errorf("invalid errors key for response profile %s: %s", rc, profile.Errors.Key)
			}
		}
	}
	if err := cache.StoreProfiles(&transformed, ctx); err != nil {
		return err
	}
	return nil
}
