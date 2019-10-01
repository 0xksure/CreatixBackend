package config

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env    string
	DbUser string
	DbPass string
	DbName string
	DbHost string
	DbPort string
}

// SetUpConfig sets up the correct configuration for the app
func SetUpConfig() (cfg *Config, err error) {

	logrus.SetFormatter(&logrus.JSONFormatter{})

	e := godotenv.Load()
	if e != nil {
		return &Config{}, err
	}

	dbUser := os.Getenv("db_user")
	if dbUser == "" {
		return &Config{}, errors.New("database user is not passed")
	}

	dbPass := os.Getenv("db_pass")
	if dbPass == "" {
		return &Config{}, errors.New("database password is not defined")
	}

	dbHost := os.Getenv("db_host")
	if dbHost == "" {
		return &Config{}, errors.New("database host is not defined")
	}

	dbName := os.Getenv("db_name")
	if dbName == "" {
		return &Config{}, errors.New("database name is not defined")
	}

	dbPort := os.Getenv("db_port")
	if dbPort == "" {
		return &Config{}, errors.New("database port is not defined ")
	}

	env := os.Getenv("env")
	if env == "" {
		log.Printf("environment not set. Setting to dev")
		env = "dev"
	}

	return &Config{
		env,
		dbUser,
		dbPass,
		dbName,
		dbHost,
		dbPort,
	}, nil
}
