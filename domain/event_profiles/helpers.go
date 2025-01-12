package eventprofiles

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

func transformProfiles(profiles *[]Profile) map[string]Profile {
	if profiles == nil {
		return nil
	}
	transformed := make(map[string]Profile, len(*profiles))
	for _, p := range *profiles {
		transformed[p.Trigger] = p
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
		for _, rc := range common.EventCodes {
			if profile, ok := transformed[rc]; !ok {
				return fmt.Errorf("response profile not found for code: %s", rc)
			} else if profile.ResponseBody == nil {
				return fmt.Errorf("invalid response body found for %s", rc)
			}
		}
	}
	if err := cache.StoreProfiles(&transformed, ctx); err != nil {
		return err
	}
	return nil
}
