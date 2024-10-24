package application

import (
	"context"
	"fmt"
	"ifttt/handler/application/config"
	"ifttt/handler/application/controllers"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var currCore *core.ServerCore

func Init() {
	if core, err := core.NewServerCore(); err != nil {
		panic(fmt.Errorf("could not create core: %s", err))
	} else {
		currCore = core
	}

	port := config.GetConfigProp("app.port")
	app := fiber.New()
	ctx := context.Background()

	if err := createApis(app, ctx); err != nil {
		panic(err)
	}

	if err := createCronJobs(ctx); err != nil {
		panic(err)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}

func createApis(fiber *fiber.App, ctx context.Context) error {
	apis, err := currCore.ConfigStore.APIPersistentRepo.GetAllApis(ctx)
	if err != nil {
		return fmt.Errorf("could not get apis from persistent config store: %s", err)
	}

	if err := currCore.CacheStore.APICacheRepo.StoreApis(apis, ctx); err != nil {
		return fmt.Errorf("could not store apis in cache storage: %s", err)
	}

	if apis != nil {
		for _, currApi := range *apis {
			if matched, err := common.RegexpArrayMatch(common.ReservedPaths, currApi.Path); err != nil {
				return err
			} else if matched {
				fmt.Printf("ServerInit: skipping api path: %s | paths not allowed: %s",
					currApi.Path, strings.Join(common.ReservedPaths, ", "))
				continue
			}
			fmt.Printf("attempting to attach %s to routes\n", currApi.Path)
			if err := controllers.NewMainController(fiber, currCore, &currApi, ctx); err != nil {
				fmt.Printf("failed to attach route %s", currApi.Path)
			}
		}
	}

	return nil
}

func createCronJobs(ctx context.Context) error {
	cronJobs, err := currCore.ConfigStore.CronPersistentRepo.GetAllCronJobs(ctx)
	if err != nil {
		return fmt.Errorf("could not get cron jobs from persistent config store: %s", err)
	}

	if err := currCore.CacheStore.CronCacheRepo.StoreCrons(cronJobs, ctx); err != nil {
		return fmt.Errorf("could not store cronjobs in cache storage: %s", err)
	}

	if cronJobs != nil {
		for _, currCron := range *cronJobs {
			fmt.Printf("attempting to attach cronjob %s\n", currCron.Name)
			if err := controllers.NewCronController(&currCron, currCore, ctx); err != nil {
				fmt.Printf("failed to attach cronjob %s", currCron.Name)
			}
		}
		currCore.Cron.Start()
	}

	return nil
}
