package application

import (
	"context"
	"fmt"
	"handler/application/config"
	"handler/application/controllers"
	"handler/application/core"
	"handler/domain/api"
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

	serializedApis, err := currCore.ConfigStore.APIPersistentRepo.GetAllApis(ctx)
	if err != nil {
		panic(fmt.Errorf("could not get apis from persistent config store: %s", err))
	}

	currCore.CacheStore.APICacheRepo.StoreApis(serializedApis, ctx)

	unserializedApis, err := api.UnserializeApis(serializedApis)
	if err != nil {
		panic(fmt.Errorf("could not get unserialize apis: %s", err))
	}

	controllers.NewTestRulesController(app, currCore)
	controllers.NewTestDumpingController(app, currCore)
	if unserializedApis != nil {
		for _, currApi := range *unserializedApis {
			if currApi.Path == controllers.TestDumpingRoute || currApi.Path == controllers.TestRulesRoute {
				fmt.Printf("ServerInit: skipping api name: %s group %s due to unusable path %s",
					currApi.Name, currApi.Group, currApi.Path)
				continue
			}
			fmt.Printf("attempting to attach %s api %s to routes\n", currApi.Type, currApi.Path)
			switch strings.ToLower(currApi.Type) {
			case api.RulesApiType:
				err = controllers.NewRulesController(app, currCore, &currApi)
			case api.DumpingApiType:
				controllers.NewDumpingController(app, currCore, &currApi)
			default:
				fmt.Printf("skipping api %s type %s not valid\n", currApi.Path, currApi.Type)
			}
		}
	}

	if err != nil {
		panic(err)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}
