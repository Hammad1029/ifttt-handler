package api

import (
	"fmt"
	"ifttt/handler/domain/configuration"
	"ifttt/handler/domain/resolvable"
)

func AttachResponseProfiles(apis *[]Api, profiles *[]configuration.ResponseProfile) error {
	transformedProfiles := configuration.TransformProfiles(profiles)

	for idx, a := range *apis {
		for event, profile := range a.Response {
			if profile.UseProfile != "" {
				if p, ok := (*transformedProfiles)[profile.UseProfile]; !ok {
					return fmt.Errorf("profile %s not found", profile.UseProfile)
				} else {
					(*apis)[idx].Response[event] = resolvable.ResponseDefinition{
						UseProfile:     p.Name,
						Definition:     p.BodyFormat,
						HTTPStatusCode: p.ResponseHTTPStatus,
					}
				}
			}
		}
	}

	return nil
}
