package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kristofhb/CreatixBackend/config"
)

var db *gorm.DB

func ConnectDB(cfg *config.Config) {

	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", &cfg.DbHost, &cfg.DbUser, cfg.DbName, cfg.DbPass)
	fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
	}

	db = conn
	db.Debug().AutoMigrate(&Account{})
}

func GetDB() *gorm.DB {
	return db
}
