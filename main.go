package main

import (
	"fmt"
	"ifttt/handler/application"
	"ifttt/handler/application/config"
)

func main() {
	fmt.Println("Starting handler")
	config.Init()
	application.Init()
}
