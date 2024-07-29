package redisUtils

import (
	"context"
	"encoding/json"
	"fmt"
	"handler/common"
	"handler/config"
	"handler/models"
	"handler/scylla"
	"strings"

	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/scylladb/gocqlx/v2/qb"
)

var redisClient *redis.Client

func Init() {
	db, err := strconv.Atoi(config.GetConfigProp("redis.db"))
	if err != nil {
		common.HandleError(err, "failed to connect to redis")
		return
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetConfigProp("redis.host"),
		Password: config.GetConfigProp("redis.password"),
		DB:       db,
	})
}

func GetRedis() *redis.Client {
	return redisClient
}

func ReadApisToRedis(ctx context.Context) {
	var apis []models.ApiModelSerialized

	stmt, names := qb.Select("apis").ToCql()
	q := scylla.GetScylla().Query(stmt, names)
	if err := q.SelectRelease(&apis); err != nil {
		common.HandleError(err)
		return
	}

	for _, v := range apis {
		deserialized, err := v.Deserialize()
		if err != nil {
			common.HandleError(err, "failed to store apis in redis")
			return
		}

		marshalled, err := json.Marshal(deserialized)
		if err != nil {
			common.HandleError(err, "failed to store apis in redis")
			return
		}
		if err := redisClient.HSet(ctx, "apis", fmt.Sprintf("%s.%s", v.ApiGroup, v.ApiName), string(marshalled)).Err(); err != nil {
			common.HandleError(err, "failed to store apis in redis")
			return
		}
	}
}

func GetAllApis() ([]*models.ApiModel, error) {
	var apis []*models.ApiModel
	apiJSON, err := GetRedis().HGetAll(context.Background(), "apis").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	for _, api := range lo.Values(apiJSON) {
		var apiUnmarshalled models.ApiModel
		err = json.Unmarshal([]byte(api), &apiUnmarshalled)
		if err != nil {
			return nil, err
		}
		apis = append(apis, &apiUnmarshalled)
	}
	return apis, nil
}

func GetApi(c *fiber.Ctx) (*models.ApiModel, error) {
	var api models.ApiModel
	apiSplit := strings.Split(c.Path(), "/")
	apiGroup := apiSplit[1]
	apiName := apiSplit[2]
	apiJSON, err := GetRedis().HGet(c.Context(), "apis", fmt.Sprintf("%s.%s", apiGroup, apiName)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(apiJSON), &api)
	if err != nil {
		return nil, err
	}
	return &api, nil
}
