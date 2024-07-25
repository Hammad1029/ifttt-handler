package main

import (
	"context"
	"fmt"
	"handler/config"
	"handler/redisUtils"
	"handler/scylla"
	"handler/server"
)

func main() {
	ctx := context.Background()

	fmt.Println("Starting handler")

	config.Init()
	scylla.Init()
	redisUtils.Init()
	redisUtils.ReadApisToRedis(ctx)

	server.Init()
}
