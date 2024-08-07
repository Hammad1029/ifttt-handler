package main

import (
	"fmt"
	"handler/app"
	"handler/config"
)

func main() {
	fmt.Println("Starting handler")
	config.Init()
	app.Init()
}
