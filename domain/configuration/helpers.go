package configuration

import (
	"context"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"reflect"
	"sync"

	"github.com/mitchellh/mapstructure"
)

func TransformProfiles(profiles *[]ResponseProfile) *map[string]ResponseProfile {
	if profiles == nil {
		return nil
	}
	transformed := make(map[string]ResponseProfile, len(*profiles))
	for _, p := range *profiles {
		transformed[p.Name] = p
	}
	return &transformed
}

func ScanToInternalTagFunc(ctx context.Context) func(tagName string, value any) error {
	reqData := request_data.GetRequestData(ctx)
	return func(tagName string, value any) error {
		if tagName != "" {
			return reqData.SetStore(tagName, value)
		}
		return nil
	}
}

func ScanFromInternalTags(format map[string]any, ctx context.Context) (*map[string]any, error) {
	wg := sync.WaitGroup{}
	cancelCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	outputMap := sync.Map{}

	for key, val := range format {
		wg.Add(1)
		go func(k string, v any) {
			defer wg.Done()
			if setVal, err := scanFromInternalTagMaybe(v, ctx); err != nil {
				cancel(err)
				return
			} else {
				select {
				case <-cancelCtx.Done():
					return
				default:
					outputMap.Store(k, setVal)
				}
			}
		}(key, val)
	}
	wg.Wait()

	if err := context.Cause(cancelCtx); err != nil {
		return nil, err
	}

	unsynced := common.SyncMapUnsync(&outputMap)
	return &unsynced, nil
}

func scanFromInternalTagMaybe(val any, ctx context.Context) (any, error) {
	reflected := reflect.Indirect(reflect.ValueOf(val))
	indirectValue := reflected.Interface()
	if reflected.Kind() == reflect.Map {
		var internalTag InternalTagInMap
		if err := mapstructure.Decode(indirectValue, &internalTag); err != nil {
			return nil, err
		} else if internalTag.InternalTag != "" {
			reqData := request_data.GetRequestData(ctx)
			if internalTag.InternalTag[0] != '.' {
				internalTag.InternalTag = "." + internalTag.InternalTag
			}
			return common.MapJqGet(&reqData.Store, internalTag.InternalTag)
		} else {
			var mapCloned map[string]any
			if err := mapstructure.Decode(indirectValue, &mapCloned); err != nil {
				return nil, err
			} else if recursedMap, err := ScanFromInternalTags(mapCloned, ctx); err != nil {
				return nil, err
			} else {
				return recursedMap, nil
			}
		}
	} else {
		return indirectValue, nil
	}
}
