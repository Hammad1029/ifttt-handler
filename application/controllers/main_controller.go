package controllers

import (
	"context"
	"fmt"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
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

func NewMainController(router fiber.Router, core *core.ServerCore, api *api.Api, ctx context.Context) error {
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

func mainController(core *core.ServerCore, parentCtx context.Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		logData := common.LogEnd{Start: time.Now()}
		requestData := request_data.RequestData{}
		requestData.Initialize()
		eventChan := make(chan resolvable.Event, 1)

		var contextState sync.Map
		contextState.Store(common.ContextLogStage, common.LogStageInitation)
		contextState.Store(common.ContextLogger, core.Logger)
		contextState.Store(common.ContextExternalExecTime, uint64(0))
		contextState.Store(common.ContextResponseSent, false)
		contextState.Store(common.ContextEventChannel, eventChan)
		contextState.Store(common.ContextRequestData, &requestData)

		cancelCtx, cancel := context.WithCancelCause(parentCtx)
		ctx := context.WithValue(cancelCtx, common.ContextState, &contextState)
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
			logData.RequestData = *requestData.UnSync()

			cancelCause := context.Cause(cancelCtx)
			if cancelCause != nil && cancelCause != context.Canceled {
				logData.Error = cancelCause.Error()
			}
			common.LogWithTracer(common.LogSystem, "request end", structs.Map(logData),
				logData.Error != "", ctx)
		}(&requestData, &logData, &contextState)

		go func() {
			if tracer, err := uuid.NewRandom(); err != nil {
				requestData.AddErrors(err)
				core.Logger.Info(fmt.Sprintf(
					"Request recieved: %s | Start time: %s", c.Path(), logData.Start.String(),
				))
				cancel(err)
				core.Logger.Error("could not assign tracer", err)
				event := &resolvable.Event{Trigger: common.EventCodes[common.EventSystemMalfunction]}
				event.ChannelSend(eventChan, ctx)
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
				event := &resolvable.Event{Trigger: common.EventCodes[common.EventNotFound]}
				event.ChannelSend(eventChan, ctx)
				return
			} else {
				logData.ApiName = api.Name
				logData.ApiPath = api.Path
				common.LogWithTracer(common.LogSystem,
					fmt.Sprintf("api found | path: %s | name: %s", api.Path, api.Name),
					api, false, ctx)
			}

			if api.Method != http.MethodGet {
				contextState.Store(common.ContextLogStage, common.LogStageParsing)
				if err := c.BodyParser(&requestData.ReqBody); err != nil {
					defer cancel(err)
					requestData.AddErrors(err)
					common.LogWithTracer(common.LogSystem, "could not parse body", err, true, ctx)
					event := &resolvable.Event{Trigger: common.EventCodes[common.EventBadRequest]}
					event.ChannelSend(eventChan, ctx)
					return
				}
				common.LogWithTracer(common.LogSystem, "request parsed", map[string]any{
					"body":    requestData.ReqBody,
					"headers": requestData.Headers,
				}, false, ctx)
			}

			for k, v := range c.GetReqHeaders() {
				requestData.Headers[k] = strings.Join(v, ",")
			}

			contextState.Store(common.ContextLogStage, common.LogStageValidation)
			if vErr := requestvalidator.ValidateMap(&api.Request, &requestData.ReqBody); len(vErr) != 0 {
				defer cancel(nil)
				err := requestvalidator.Normalize(vErr)
				requestData.AddErrors(err...)
				common.LogWithTracer(common.LogSystem, "request validation failed", vErr, false, ctx)
				event := &resolvable.Event{Trigger: common.EventCodes[common.EventBadRequest]}
				event.ChannelSend(eventChan, ctx)
				return
			} else {
				common.LogWithTracer(common.LogSystem, "request validation passed", nil, false, ctx)
			}

			contextState.Store(common.ContextLogStage, common.LogStagePreConfig)
			if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
				defer cancel(err)
				requestData.AddErrors(err)
				common.LogWithTracer(common.LogSystem, "could not prepare pre config", err, true, ctx)
				event := &resolvable.Event{Trigger: common.EventCodes[common.EventSystemMalfunction]}
				event.ChannelSend(eventChan, ctx)
				return
			} else {
				common.LogWithTracer(common.LogSystem, "pre config resolution done", nil, false, ctx)
			}

			contextState.Store(common.ContextLogStage, common.LogStageExecution)
			if err := core.InitExecution(api.Triggers, ctx); err != nil {
				cancel(err)
			}
			eventTrigger := common.EventCodes[common.EventExhaust]
			err = context.Cause(cancelCtx)
			if err != nil {
				eventTrigger = common.EventCodes[common.EventSystemMalfunction]
				requestData.AddErrors(err)
			}
			event := &resolvable.Event{Trigger: eventTrigger}
			event.ChannelSend(eventChan, ctx)
		}()

		eventRes := <-eventChan
		if response, status, err := eventRes.HandlerTrigger(ctx, c.Context(), core.ResolvableDependencies); err != nil {
			return c.Status(status).JSON(common.ResponseDefaultMalfunction)
		} else {
			return c.Status(status).JSON(response)
		}
	}
}
