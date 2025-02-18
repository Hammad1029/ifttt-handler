package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/configuration"
)

type Response struct {
	Event uint `json:"event" mapstructure:"event"`
}

type ResponseDefinition struct {
	UseProfile     string         `json:"useProfile" mapstructure:"useProfile"`
	Definition     map[string]any `json:"definition" mapstructure:"definition"`
	HTTPStatusCode int            `json:"httpStatusCode" mapstructure:"httpStatusCode"`
}

func (e *Response) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	requestState := common.GetCtxState(ctx)

	if channelUncasted, ok := requestState.Load(common.ContextResponseChannel); !ok {
		return nil, fmt.Errorf("response channel not found")
	} else if channel, ok := channelUncasted.(chan Response); !ok {
		return nil, fmt.Errorf("response channel type assertion failed")
	} else {
		e.ChannelSend(channel, ctx)
	}

	return nil, nil
}

func (e *Response) ChannelSend(responseChannel chan Response, ctx context.Context) {
	if ok := common.SetResponseSent(ctx); ok {
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("Sending response with event %d", e.Event), e, false, ctx)
		responseChannel <- *e
		close(responseChannel)
	}
}

func (e *Response) HandlerEvent(ctx context.Context, dependencies map[common.IntIota]any) (*map[string]any, int, error) {
	apiProfilesUncasted, ok := common.GetCtxState(ctx).Load(common.ContextResponseProfiles)
	if !ok {
		return nil, 0, fmt.Errorf("no api profiles found")
	}
	apiProfiles, ok := apiProfilesUncasted.(map[uint]ResponseDefinition)
	if !ok {
		return nil, 0, fmt.Errorf("could not cast response profiles")
	}

	responseDefinition, ok := apiProfiles[e.Event]
	if !ok {
		return nil, 0, fmt.Errorf("response definition for event %d not found", e.Event)
	}

	responseScanned, err := configuration.ScanFromInternalTags(responseDefinition.Definition, ctx)
	if err != nil {
		return nil, 0, err
	}

	return responseScanned, responseDefinition.HTTPStatusCode, nil
}
