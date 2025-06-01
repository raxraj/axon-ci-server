package main

import (
	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/config"
	"github.com/raxraj/axon-ci-server/routes"
	"github.com/spf13/viper"
	"log"
)

func main() {
	config.InitConfig()
	e := echo.New()
	routes.Routes(e)
	err := e.Start(":" + viper.GetString("port"))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
