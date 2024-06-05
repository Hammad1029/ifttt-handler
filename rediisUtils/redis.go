package redisUtils

import (
	"context"
	"encoding/json"
	"handler/config"
	"handler/models"
	"handler/scylla"
	"handler/utils"

	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/scylladb/gocqlx/v2/qb"
)

var redisClient *redis.Client

func Init() {
	db, error := strconv.Atoi(config.GetConfigProp("redis.db"))
	if error != nil {
		utils.HandleError(error)
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
	var apis []models.ApiModel

	stmt, names := qb.Select("apis").ToCql()
	q := scylla.GetScylla().Query(stmt, names)
	if err := q.SelectRelease(&apis); err != nil {
		utils.HandleError(err)
		return
	}

	for _, v := range apis {
		jsonData, err := json.Marshal(v)
		if err != nil {
			utils.HandleError(err)
			return
		}
		if err := redisClient.HSet(ctx, "apis", v.ApiName, jsonData).Err(); err != nil {
			utils.HandleError(err)
			return
		}
	}

}
