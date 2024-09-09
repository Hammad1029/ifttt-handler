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

	apis, err := currCore.ConfigStore.APIPersistentRepo.GetAllApis(ctx)
	if err != nil {
		panic(fmt.Errorf("could not get apis from persistent config store: %s", err))
	}

	if err := currCore.CacheStore.APICacheRepo.StoreApis(apis, ctx); err != nil {
		panic(fmt.Errorf("could not store apis in cache storage"))
	}

	// controllers.NewTestRulesController(app, currCore)
	// controllers.NewTestDumpingController(app, currCore)
	if apis != nil {
		for _, currApi := range *apis {
			if matched, err := common.RegexpArrayMatch(common.ReservedPaths, currApi.Path); err != nil {
				panic(err)
			} else if matched {
				fmt.Printf("ServerInit: skipping api path: %s | paths not allowed: %s",
					currApi.Path, strings.Join(common.ReservedPaths, ", "))
				continue
			}
			fmt.Printf("attempting to attach %s to routes\n", currApi.Path)
			err = controllers.NewRulesController(app, currCore, &currApi)
		}
	}

	if err != nil {
		panic(err)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}
