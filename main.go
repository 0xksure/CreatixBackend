package main

import (
	"log"

	"github.com/kristofhb/CreatixBackend/app"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/logging"
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

	a, err := app.New(cfg)
	if err != nil {
		standardLogger.Error(err)
	}

	a.Run()
}
