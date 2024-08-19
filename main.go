package main

import (
	"fmt"
	"handler/application"
	"handler/application/config"
)

func main() {
	fmt.Println("Starting handler")
	config.Init()
	application.Init()
}
