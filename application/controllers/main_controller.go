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
		resChan := make(chan resolvable.ResponseResolvable, 1)

		var contextState sync.Map
		contextState.Store(common.ContextLogStage, common.LogStageInitation)
		contextState.Store(common.ContextLogger, core.Logger)
		contextState.Store(common.ContextExternalExecTime, uint64(0))
		contextState.Store(common.ContextResponseSent, false)
		contextState.Store(common.ContextResponseChannel, resChan)
		contextState.Store(common.ContextRequestData, &requestData)

		cancelCtx, cancel := context.WithCancelCause(parentCtx)
		ctx := context.WithValue(cancelCtx, common.ContextState, &contextState)

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

		tracer, err := uuid.NewRandom()
		if err != nil {
			cancel(err)
			core.Logger.Info(fmt.Sprintf(
				"Request recieved: %s | Start time: %s", c.Path(), logData.Start.String(),
			))
			core.Logger.Error("could not assign tracer", err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not assign tracer",
			}
			res.ManualSend(resChan, err, ctx)
			return c.JSON(res)
		} else {
			contextState.Store(common.ContextTracer, tracer.String())
			common.LogWithTracer(common.LogSystem, fmt.Sprintf(
				"Request recieved: %s | Start time: %s", c.Path(), logData.Start.String(),
			), nil, false, ctx)
		}

		contextState.Store(common.ContextLogStage, common.LogStageMemload)
		api, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if api == nil || err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem,
				fmt.Sprintf("api not found | path: %s", c.Path()), err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "404",
				ResponseDescription: "API not found",
			}
			res.ManualSend(resChan, err, ctx)
			return c.JSON(res)
		} else {
			logData.ApiName = api.Name
			logData.ApiPath = api.Path
			common.LogWithTracer(common.LogSystem,
				fmt.Sprintf("api found | path: %s | name: %s", api.Path, api.Name),
				api, false, ctx)
		}

		contextState.Store(common.ContextLogStage, common.LogStageParsing)
		if err := c.BodyParser(&requestData.ReqBody); err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem, "could not parse body", err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Error in parsing body",
			}
			res.ManualSend(resChan, err, ctx)
			return c.JSON(res)
		}
		requestData.Headers = c.GetReqHeaders()
		common.LogWithTracer(common.LogSystem, "request parsed", map[string]any{
			"body":    requestData.ReqBody,
			"headers": requestData.Headers,
		}, false, ctx)

		contextState.Store(common.ContextLogStage, common.LogStageValidation)
		if err := requestvalidator.ValidateMap(&api.Request, &requestData.ReqBody); len(err) != 0 {
			defer cancel(nil)
			common.LogWithTracer(common.LogSystem, "request validation failed", err, false, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Validation error",
			}
			res.AddValidationErrors(err)
			res.ManualSend(resChan, nil, ctx)
			return c.JSON(res)
		} else {
			common.LogWithTracer(common.LogSystem, "request validation passed", nil, false, ctx)
		}

		contextState.Store(common.ContextLogStage, common.LogStagePreConfig)
		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem, "could not prepare pre config", err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			res.ManualSend(resChan, err, ctx)
			return c.JSON(res)
		} else {
			common.LogWithTracer(common.LogSystem, "pre config resolution done", nil, false, ctx)
		}

		go func() {
			defer cancel(nil)

			contextState.Store(common.ContextLogStage, common.LogStagePreWare)
			if err := core.InitMiddleWare(api.PreWare, ctx); err != nil {
				cancel(err)
			} else {
				contextState.Store(common.ContextLogStage, common.LogStageMainWare)
				if err := core.InitMainWare(api.MainWare, ctx); err != nil {
					cancel(err)
				} else {
					contextState.Store(common.ContextLogStage, common.LogStagePostWare)
					if err := core.InitMiddleWare(api.PostWare, ctx); err != nil {
						cancel(err)
					}
				}
			}

			var res resolvable.ResponseResolvable
			err := context.Cause(cancelCtx)
			if err != nil {
				res = resolvable.ResponseResolvable{
					ResponseCode:        common.ResponseCodeSystemError,
					ResponseDescription: common.ResponseDescriptionSystemError,
				}
			}
			res.ManualSend(resChan, err, ctx)
		}()

		res := <-resChan
		return c.JSON(res)
	}
}
