package application

import (
	"context"
	"fmt"
	"ifttt/handler/application/config"
	"ifttt/handler/common"
	"ifttt/handler/domain/api"
	"ifttt/handler/domain/orm_schema"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

var currCore *ServerCore

func Init() {
	if core, err := newServerCore(); err != nil {
		panic(fmt.Errorf("could not create core: %s", err))
	} else {
		currCore = core
	}

	port := config.GetConfigProp("app.port")
	app := fiber.New()
	app.Use(pprof.New())
	ctx := context.Background()

	currCore.Logger.Info("creating APIs")
	if err := createApis(app, ctx); err != nil {
		panic(err)
	}

	currCore.Logger.Info("getting and storing models")
	if err := orm_schema.GetAndStoreModels(
		currCore.ConfigStore.OrmRepo, currCore.CacheStore.OrmRepo, ctx,
	); err != nil {
		panic(err)
	}

	currCore.Logger.Info("getting and storing associations")
	if err := orm_schema.GetAndStoreAssociations(
		currCore.ConfigStore.OrmRepo, currCore.CacheStore.OrmRepo, ctx,
	); err != nil {
		panic(err)
	}

	currCore.Logger.Info("creating Cron Jobs")
	if err := createCronJobs(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		listener, err := net.Listen("unix", common.SelfSocket)
		if err != nil {
			panic(err)
		}
		defer os.Remove(common.SelfSocket)

		if err := os.Chmod(common.SelfSocket, 0666); err != nil {
			panic(err)
		}

		fmt.Printf("Server running on unix socket: %s (for cronjobs) \n", common.SelfSocket)
		if err := app.Listener(listener); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
			panic(err)
		}
		fmt.Printf("Handler running on port: %s \n", port)
	}()

	wg.Wait()
}

func createApis(fiber *fiber.App, ctx context.Context) error {
	apis, err := currCore.ConfigStore.APIRepo.GetAllApis(ctx)
	if err != nil {
		return fmt.Errorf("could not get apis from persistent config store: %s", err)
	}

	profiles, err := currCore.ConfigStore.ResponseProfileRepo.GetAllProfiles()
	if err != nil {
		return fmt.Errorf("could not get response profiles: %s", err)
	}

	if err := api.AttachResponseProfiles(apis, profiles); err != nil {
		return fmt.Errorf("could not attach response profiles to apis: %s", err)
	}

	if err := currCore.CacheStore.APIRepo.StoreApis(apis, ctx); err != nil {
		return fmt.Errorf("could not store apis in cache storage: %s", err)
	}

	if apis != nil {
		for _, currApi := range *apis {
			if matched, err := common.RegexpArrayMatch(common.ReservedPaths, currApi.Path); err != nil {
				return err
			} else if matched {
				fmt.Printf("skipping api path: %s | paths not allowed: %s",
					currApi.Path, strings.Join(common.ReservedPaths, ", "))
				continue
			}
			fmt.Printf("attempting to attach %s to routes\n", currApi.Path)
			if err := newMainController(fiber, currCore, &currApi, ctx); err != nil {
				fmt.Printf("failed to attach route %s", currApi.Path)
			}
		}
	}

	return nil
}

func createCronJobs(ctx context.Context) error {
	os.Remove(common.SelfSocket)

	cronJobs, err := currCore.ConfigStore.CronRepo.GetAllCronJobs(ctx)
	if err != nil {
		return fmt.Errorf("could not get cron jobs from persistent config store: %s", err)
	}

	if cronJobs != nil {
		for _, currCron := range *cronJobs {
			fmt.Printf("attempting to attach cronjob %s\n", currCron.Name)
			if err := currCore.addCronJob(&currCron); err != nil {
				fmt.Printf("failed to attach cronjob %s", currCron.Name)
			}
		}
		currCore.Cron.Start()
	}

	return nil
}
