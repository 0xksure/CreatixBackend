package models

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

/*
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
	fmt.Println("Auto Migrate")
	db.Debug().AutoMigrate(&Account{})
	db.Debug().AutoMigrate(&Newsletter{})
}

func GetDB() *gorm.DB {
	return db
}
*/
