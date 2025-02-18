package application

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/configuration"
	"ifttt/handler/domain/request_data"
	requestvalidator "ifttt/handler/domain/request_validator.go"
	"ifttt/handler/domain/resolvable"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func newMainController(router fiber.Router, core *ServerCore, api *api.Api, ctx context.Context) error {
	controller := mainController(core, ctx)
	switch strings.ToUpper(api.Method) {
	case http.MethodGet:
		router.Get(api.Path, controller)
	case http.MethodPost:
		router.Post(api.Path, controller)
	case http.MethodPut:
		router.Put(api.Path, controller)
	case http.MethodDelete:
		router.Delete(api.Path, controller)
	default:
		return fmt.Errorf("method NewMainController: method %s not found", api.Method)
	}
	return nil
}

func mainController(core *ServerCore, parentCtx context.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		logData := common.LogEnd{Start: time.Now()}
		requestData := request_data.NewRequestData()
		responseChan := make(chan resolvable.Response, 1)

		var contextState sync.Map
		contextState.Store(common.ContextLogStage, common.LogStageInitation)
		contextState.Store(common.ContextLogger, core.Logger)
		contextState.Store(common.ContextExternalExecTime, uint64(0))
		contextState.Store(common.ContextResponseSent, false)
		contextState.Store(common.ContextResponseChannel, responseChan)
		contextState.Store(common.ContextRequestData, requestData)

		valueCtx := context.WithValue(parentCtx, common.ContextState, &contextState)
		cancelCtx, cancel := context.WithCancelCause(valueCtx)
		defer cancel(nil)

		go func(requestData *request_data.RequestData, logData *common.LogEnd, ctxState *sync.Map) {
			<-cancelCtx.Done()
			ctxState.Store(common.ContextLogStage, common.LogStageEnding)

			logData.End = time.Now()
			logData.ExecutionTime = uint64(logData.End.Sub(logData.Start).Milliseconds())
			externalExecTime, ok := ctxState.Load(common.ContextExternalExecTime)
			if !ok {
				return
			}
			logData.ExternalExecTime = externalExecTime.(uint64)
			logData.InternalExecTime = logData.ExecutionTime - logData.ExternalExecTime
			logData.RequestData = structs.Map(requestData)

			cancelCause := context.Cause(valueCtx)
			if cancelCause != nil && cancelCause != context.Canceled {
				logData.Error = cancelCause.Error()
			}
			common.LogWithTracer(common.LogSystem, "request end", structs.Map(logData),
				logData.Error != "", cancelCtx)
		}(requestData, &logData, &contextState)

		go func(ctx context.Context) {
			if tracer, err := uuid.NewRandom(); err != nil {
				requestData.AddErrors(err)
				core.Logger.Info(fmt.Sprintf(
					"Request recieved: %s | Start time: %s", c.Path(), logData.Start.String(),
				))
				cancel(err)
				core.Logger.Error("could not assign tracer", err)
				res := &resolvable.Response{Event: common.EventCodes[common.EventSystemMalfunction]}
				res.ChannelSend(responseChan, ctx)
				return
			} else {
				tracerStr := tracer.String()
				c.Set(common.ResponseHeaderTracer, tracerStr)
				c.Set(common.ResponseHeaderContentType, "application/json")
				contextState.Store(common.ContextTracer, tracerStr)
				common.LogWithTracer(common.LogSystem, fmt.Sprintf(
					"Request recieved: %s | Start time: %s", c.Path(), logData.Start.String(),
				), nil, false, ctx)
			}

			contextState.Store(common.ContextLogStage, common.LogStageMemload)
			api, err := core.CacheStore.APIRepo.GetApiByPath(c.Path(), ctx)
			if api == nil || err != nil {
				requestData.AddErrors(err)
				defer cancel(err)
				common.LogWithTracer(common.LogSystem,
					fmt.Sprintf("api not found | path: %s", c.Path()), err, true, ctx)
				response := &resolvable.Response{Event: common.EventCodes[common.EventNotFound]}
				response.ChannelSend(responseChan, ctx)
				return
			} else {
				contextState.Store(common.ContextResponseProfiles, api.Response)
				logData.ApiName = api.Name
				logData.ApiPath = api.Path
				common.LogWithTracer(common.LogSystem,
					fmt.Sprintf("api found | path: %s | name: %s", api.Path, api.Name),
					api, false, ctx)
			}

			for k, v := range c.GetReqHeaders() {
				requestData.Headers[k] = strings.Join(v, ",")
			}

			if api.Method != http.MethodGet {
				scanToInternal := configuration.ScanToInternalTagFunc(ctx)

				contextState.Store(common.ContextLogStage, common.LogStageParsing)
				reqBody, err := common.BodyParser(c)
				if err != nil {
					defer cancel(err)
					requestData.AddErrors(err)
					scanToInternal(common.InternalTagErrorValidation, err.Error())
					common.LogWithTracer(common.LogSystem, "could not parse body", err, true, ctx)
					response := &resolvable.Response{Event: common.EventCodes[common.EventBadRequest]}
					response.ChannelSend(responseChan, ctx)
					return
				}
				common.LogWithTracer(common.LogSystem, "request parsed", map[string]any{
					"body":    reqBody,
					"headers": requestData.Headers,
				}, false, ctx)

				contextState.Store(common.ContextLogStage, common.LogStageValidation)
				if vErr := requestvalidator.ValidateMap(&api.Request, reqBody, scanToInternal); len(vErr) != 0 {
					defer cancel(nil)
					err := requestvalidator.Normalize(vErr)
					requestData.AddErrors(err...)
					strErr := make([]string, 0, len(err))
					for _, e := range err {
						if str := e.Error(); str != "" {
							strErr = append(strErr, str)
						}
					}
					scanToInternal(common.InternalTagErrorValidation, strErr)
					common.LogWithTracer(common.LogSystem, "request validation failed", vErr, false, ctx)
					response := &resolvable.Response{Event: common.EventCodes[common.EventBadRequest]}
					response.ChannelSend(responseChan, ctx)
					return
				} else {
					common.LogWithTracer(common.LogSystem, "request validation passed", nil, false, ctx)
				}
			}

			select {
			case <-ctx.Done():
			default:
				contextState.Store(common.ContextLogStage, common.LogStagePreConfig)
				if _, err := resolvable.ResolveArrayMust(&api.PreConfig, ctx, core.ResolvableDependencies); err != nil {
					cancel(err)
				}
			}

			select {
			case <-ctx.Done():
			default:
				contextState.Store(common.ContextLogStage, common.LogStageExecution)
				if err := core.initExecution(api.Triggers, ctx); err != nil {
					cancel(err)
				}
			}

			responseEvent := common.EventCodes[common.EventExhaust]
			err = context.Cause(ctx)
			if err != nil {
				responseEvent = common.EventCodes[common.EventSystemMalfunction]
				requestData.AddErrors(err)
				requestData.SetStore(common.InternalTagErrorSystem, err.Error())
			}
			response := &resolvable.Response{Event: responseEvent}
			response.ChannelSend(responseChan, ctx)
		}(cancelCtx)

		res := <-responseChan
		if response, status, err := res.HandlerEvent(valueCtx, core.ResolvableDependencies); err != nil {
			return c.Status(status).JSON(common.ResponseDefaultMalfunction)
		} else {
			return c.Status(status).JSON(response)
		}
	}
}
