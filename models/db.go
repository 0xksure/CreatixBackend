package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/logging"
)

var db *gorm.DB

func ConnectDB(cfg *config.Config, stdLog *logging.StandardLogger) {
	dbUri := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbName, cfg.DbPass)
	fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
		stdLog.Unsuccessful("Not able to open postgres", err)
	}

	db = conn
	db.Debug().AutoMigrate(&Account{})
}

func GetDB() *gorm.DB {
	return db
}
