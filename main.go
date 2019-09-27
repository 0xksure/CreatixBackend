package main

import (
	"log"

	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/models"
)

//main

func main() {

	// Set up config
	cfg, err := config.SetUpConfig()
	if err != nil {
		log.Fatal("Not able to set config ")
		return
	}
	models.ConnectDB(cfg)

	handler.Restapi()
}
