package main

import (
	"context"
	"fmt"
	"handler/config"
	"handler/redisUtils"
	"handler/server"
	"handler/store"
	"os"
)

func main() {
	ctx := context.Background()

	fmt.Println("Starting handler")

	config.Init()

	if err := store.InitConfigStore(); err != nil {
		fmt.Printf("error in instantiating config store: %s \n exiting", err)
		os.Exit(1)
	}
	if err := store.InitDataStore(); err != nil {
		fmt.Printf("error in instantiating config store: %s \n exiting", err)
		os.Exit(1)
	}

	redisUtils.Init()
	redisUtils.ReadApisToRedis(ctx)

	server.Init()
}
