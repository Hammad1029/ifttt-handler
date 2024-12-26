package application

import (
	"context"
	"fmt"
	"ifttt/handler/application/config"
	"ifttt/handler/application/controllers"
	"ifttt/handler/application/core"
	"ifttt/handler/common"
	"ifttt/handler/domain/orm_schema"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
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

	if err := getAndStoreORMSchemas(ctx); err != nil {
		panic(err)
	}

	app.Listen(fmt.Sprintf(":%s", port))
	fmt.Printf("Handler running on port: %s", port)
}

func createApis(fiber *fiber.App, ctx context.Context) error {
	apis, err := currCore.ConfigStore.APIRepo.GetAllApis(ctx)
	if err != nil {
		return fmt.Errorf("could not get apis from persistent config store: %s", err)
	}

	if err := currCore.CacheStore.APIRepo.StoreApis(apis, ctx); err != nil {
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
	cronJobs, err := currCore.ConfigStore.CronRepo.GetAllCronJobs(ctx)
	if err != nil {
		return fmt.Errorf("could not get cron jobs from persistent config store: %s", err)
	}

	if err := currCore.CacheStore.CronRepo.StoreCrons(cronJobs, ctx); err != nil {
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

func getAndStoreORMSchemas(ctx context.Context) error {
	var schemas []orm_schema.Schema

	tableNames, err := currCore.DataStore.OrmSchemaRepo.GetTableNames()
	if err != nil {
		return err
	}
	columns, err := currCore.DataStore.OrmSchemaRepo.GetAllColumns(tableNames)
	if err != nil {
		return err
	}
	constraints, err := currCore.DataStore.OrmSchemaRepo.GetAllConstraints(tableNames)
	if err != nil {
		return err
	}

	groupedColumns := lo.GroupBy(*columns, func(col orm_schema.Column) string {
		return col.TableName
	})
	groupedConstraints := lo.GroupBy(*constraints, func(constraint orm_schema.Constraint) string {
		return constraint.TableName
	})

	var newSchema orm_schema.Schema
	for _, tableName := range tableNames {
		newSchema.TableName = tableName
		if columns, ok := groupedColumns[newSchema.TableName]; ok {
			newSchema.Columns = columns
		}
		if constraints, ok := groupedConstraints[newSchema.TableName]; ok {
			newSchema.Constraints = constraints
		}
		schemas = append(schemas, newSchema)
	}

	if err := currCore.CacheStore.OrmSchemaRepo.StoreSchema(&schemas, ctx); err != nil {
		return err
	}

	return nil
}
