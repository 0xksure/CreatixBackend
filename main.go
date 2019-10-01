package main

import (
	"log"

	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
)

//main

func main() {

	standardLogger := logging.NewLogger()
	// Set up config
	cfg, err := config.SetUpConfig()
	if err != nil {
		log.Fatal("Not able to set config ")
		standardLogger.Misconfigured("Configuration is misconfigured", err)
		return
	}
	models.ConnectDB(cfg, standardLogger)

	handler.Restapi(standardLogger)
}
