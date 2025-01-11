package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	responseprofiles "ifttt/handler/domain/response_profiles"

	"github.com/fatih/structs"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	ResponseCode        any      `json:"responseCode" mapstructure:"responseCode"`
	ResponseDescription string   `json:"responseDescription" mapstructure:"responseDescription"`
	Data                any      `json:"data" mapstructure:"data"`
	Errors              []string `json:"errors" mapstructure:"errors"`
}

func (r *Response) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetCtxState(ctx)
	reqData := GetRequestData(ctx)
	r.Data = common.SyncMapUnsync(reqData.Response)

	if resChanUncasted, ok := requestState.Load(common.ContextResponseChannel); !ok {
		return nil, fmt.Errorf("response channel not found")
	} else if responseChannel, ok := resChanUncasted.(chan Response); !ok {
		return nil, fmt.Errorf("method Resolve: response channel type assertion failed")
	} else {
		r.channelSend(responseChannel, ctx)
	}

	return nil, nil
}

func (r *Response) ManualSend(resChan chan Response, pErr error, ctx context.Context) {
	if !common.GetResponseSent(ctx) {
		if pErr != nil {
			r.AddError(pErr)
		}
		if _, err := r.Resolve(ctx, nil); err != nil {
			r.AddError(err)
			common.LogWithTracer(common.LogSystem, "error in resolving response", err, true, ctx)
			r.ResponseCode = common.ResponseCodes[common.ResponseCodeSystemMalfunction]
			r.channelSend(resChan, ctx)
		}
	}
}

func (r *Response) channelSend(resChan chan Response, ctx context.Context) {
	if ok := common.SetResponseSent(ctx); ok {
		if reqData := GetRequestData(ctx); reqData != nil {
			reqData.AggregatedResponse = structs.Map(r)
		}
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("Sending response | Response Code: %s | Response Description: %s",
				r.ResponseCode, r.ResponseDescription), r, false, ctx)
		resChan <- *r
		close(resChan)
	}
}

func (r *Response) AddError(err error) {
	r.safeInitialize()
	r.Errors = append(r.Errors, err.Error())
}

func (r *Response) safeInitialize() {
	if r.ResponseCode == "" || r.ResponseDescription == "" {
		r.ResponseCode = common.ResponseCodes[common.ResponseCodeSuccess]
	}
	if r.Errors == nil {
		r.Errors = make([]string, 0)
	}
}

func (r *Response) generateJSON(ctx context.Context, dependencies map[common.IntIota]any) (map[string]any, int, error) {
	profileRepo, ok := dependencies[common.DependencyResponseProfileCacheRepo].(responseprofiles.CacheRepository)
	if !ok {
		return nil, 0, fmt.Errorf("could not get profile repo")
	}

	var assumeResponse string
	if r.ResponseCode == common.ResponseCodes[common.ResponseCodeExhaust] ||
		r.ResponseCode == common.ResponseCodes[common.ResponseCodeBadRequest] ||
		r.ResponseCode == common.ResponseCodes[common.ResponseCodeSystemMalfunction] {
		assumeResponse = fmt.Sprint(r.ResponseCode)
	} else {
		assumeResponse = common.ResponseCodes[common.ResponseCodeSuccess]
	}
	internalProfile, err := profileRepo.GetProfileByCode(assumeResponse, ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("error in getting response profile")
	} else if internalProfile == nil {
		return nil, 0, fmt.Errorf("no profile found")
	}

	var mapperProfile *responseprofiles.Profile
	if internalProfile.MappedProfile != nil {
		mapperProfile = internalProfile.MappedProfile
	} else {
		mapperProfile = internalProfile
	}
	mapped := r.mapToProfile(mapperProfile)

	return mapped, mapperProfile.HttpStatus, nil
}

func (r *Response) mapToProfile(profile *responseprofiles.Profile) map[string]any {
	mapped := make(map[string]any, 4)
	profile.Code.AddToMap(r.ResponseCode, &mapped)
	profile.Code.AddToMap(r.ResponseDescription, &mapped)
	profile.Code.AddToMap(r.Data, &mapped)
	profile.Code.AddToMap(r.Errors, &mapped)
	return mapped
}

func (r *Response) FiberReturn(c *fiber.Ctx, ctx context.Context, dependencies map[common.IntIota]any) error {
	jsonMapped, status, err := r.generateJSON(ctx, dependencies)
	if err != nil {
		return c.Status(status).Send([]byte("could not map response"))
	}
	return c.Status(status).JSON(jsonMapped)
}
