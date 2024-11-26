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
		startTime := time.Now()

		var contextState sync.Map
		ctx, cancel := context.WithCancelCause(parentCtx)
		defer cancel(nil)
		ctx = context.WithValue(ctx, common.ContextState, &contextState)

		tracer, err := uuid.NewRandom()
		if err != nil {
			core.Logger.Info(fmt.Sprintf(
				"Request recieved: %s | Start time: %s", c.Path(), startTime.String(),
			))
			core.Logger.Error("could not assign tracer", err)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not assign tracer",
			}
			res.AddError(err)
			return c.JSON(res)
		}

		contextState.Store(common.ContextLogger, core.Logger)
		contextState.Store(common.ContextTracer, tracer.String())
		common.LogWithTracer(common.LogSystem, fmt.Sprintf(
			"Request recieved: %s | Start time: %s", c.Path(), startTime.String(),
		), nil, false, ctx)

		requestData := request_data.RequestData{}
		requestData.Initialize()
		resChan := make(chan resolvable.ResponseResolvable, 1)

		contextState.Store(common.ContextResponseChannel, resChan)
		contextState.Store(common.ContextExternalExecTime, uint64(0))
		contextState.Store(common.ContextRequestData, &requestData)
		contextState.Store(common.ContextResponseSent, false)

		go func(requestData *request_data.RequestData) {
			<-ctx.Done()
			end := time.Now()
			executionTime := uint64(end.Sub(startTime).Milliseconds())
			externalExecTime, ok := contextState.Load(common.ContextExternalExecTime)
			if !ok {
				return
			}
			internalExecTime := executionTime - externalExecTime.(uint64)
			common.LogWithTracer(common.LogSystem, "request end", map[string]any{
				"start":            startTime,
				"end":              end,
				"executionTime":    executionTime,
				"internalExecTime": internalExecTime,
				"externalExecTime": externalExecTime,
				"requestData":      requestData,
			}, false, ctx)
		}(&requestData)

		api, err := core.CacheStore.APICacheRepo.GetApiByPath(c.Path(), ctx)
		if api == nil || err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem,
				fmt.Sprintf("api not found | path: %s", c.Path()), err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "404",
				ResponseDescription: "API not found",
			}
			res.ManualSend(resChan, core.ResolvableDependencies, ctx)
			return c.JSON(res)
		}
		common.LogWithTracer(common.LogSystem,
			fmt.Sprintf("api found | path: %s | name: %s", api.Path, api.Name),
			api, false, ctx)

		if err := c.BodyParser(&requestData.ReqBody); err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem, "could not parse body", err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Error in parsing body",
			}
			res.ManualSend(resChan, core.ResolvableDependencies, ctx)
			return c.JSON(res)
		}
		requestData.Headers = c.GetReqHeaders()
		common.LogWithTracer(common.LogSystem, "request parsed", map[string]any{
			"body":    requestData.ReqBody,
			"headers": requestData.Headers,
		}, false, ctx)

		if err := requestvalidator.ValidateMap(&api.Request, &requestData.ReqBody); len(err) != 0 {
			defer cancel(nil)
			common.LogWithTracer(common.LogSystem, "request validation failed", err, false, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "400",
				ResponseDescription: "Validation error",
			}
			res.AddValidationErrors(err)
			res.ManualSend(resChan, core.ResolvableDependencies, ctx)
			return c.JSON(res)
		}

		if err := core.PreparePreConfig(api.PreConfig, ctx); err != nil {
			defer cancel(err)
			common.LogWithTracer(common.LogSystem, "could not prepare pre config", err, true, ctx)
			res := &resolvable.ResponseResolvable{
				ResponseCode:        "500",
				ResponseDescription: "Could not prepare pre config",
			}
			res.ManualSend(resChan, core.ResolvableDependencies, ctx)
			return c.JSON(res)
		}

		go func() {
			defer cancel(nil)

			if err := core.InitMiddleWare(api.PreWare, ctx); err != nil {
				cancel(err)
			} else if err := core.InitMainWare(api.MainWare, ctx); err != nil {
				cancel(err)
			} else if err := core.InitMiddleWare(api.PostWare, ctx); err != nil {
				cancel(err)
			}

			var res resolvable.ResponseResolvable
			if err := context.Cause(ctx); err != nil {
				res = resolvable.ResponseResolvable{
					ResponseCode:        common.ResponseCodeSystemError,
					ResponseDescription: common.ResponseDescriptionSystemError,
				}
				res.AddError(err)
			}
			res.ManualSend(resChan, core.ResolvableDependencies, ctx)
		}()

		res := <-resChan
		return c.JSON(res)
	}
}
