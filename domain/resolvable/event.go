package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	eventprofiles "ifttt/handler/domain/event_profiles"
)

type Event struct {
	Trigger string `json:"trigger" mapstructure:"trigger"`
}

func (e *Event) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetCtxState(ctx)

	if eventChanUncasted, ok := requestState.Load(common.ContextEventChannel); !ok {
		return nil, fmt.Errorf("event channel not found")
	} else if eventChannel, ok := eventChanUncasted.(chan Event); !ok {
		return nil, fmt.Errorf("event channel type assertion failed")
	} else {
		e.ChannelSend(eventChannel, ctx)
	}

	return nil, nil
}

func (e *Event) ChannelSend(eventChan chan Event, ctx context.Context) {
	if ok := common.SetResponseSent(ctx); ok {
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("Sending event with trigger %s", e.Trigger), e, false, ctx)
		eventChan <- *e
		close(eventChan)
	}
}

func (e *Event) HandlerTrigger(ctx context.Context, dependencies map[common.IntIota]any) (map[string]any, int, error) {
	profileRepo, ok := dependencies[common.DependencyEventProfileCacheRepo].(eventprofiles.CacheRepository)
	if !ok {
		return nil, 0, fmt.Errorf("could not get event profile repo")
	}

	var useProfile *eventprofiles.Profile
	if internalProfile, err := profileRepo.GetProfileByTrigger(e.Trigger, ctx); err != nil {
		return nil, 0, err
	} else if internalProfile == nil {
		return nil, 0, fmt.Errorf("profile for trigger %s not found", e.Trigger)
	} else if internalProfile.MappedProfiles != nil && len(*internalProfile.MappedProfiles) > 0 {
		useProfile = &(*internalProfile.MappedProfiles)[0]
	} else {
		useProfile = internalProfile
	}

	var response map[string]any
	reqData := GetRequestData(ctx)
	if !useProfile.UseBody {
		response = reqData.Response
	} else if resolved, err := resolveMapMaybeParallel(&useProfile.ResponseBody, ctx, dependencies); err != nil {
		return nil, useProfile.ResponseHTTPStatus, err
	} else {
		response = resolved
	}

	return response, useProfile.ResponseHTTPStatus, nil
}
