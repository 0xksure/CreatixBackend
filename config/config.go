package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env          string `default:""`
	DatabaseURL  string `default:"" envconfig:"DATABASE_URL"`
	Port         string `default:":8080"`
	SMTPServer   string `default:""`
	SMTPPWD      string `default:""`
	SMTPUserName string `default:""`
	JwtSecret    string `default:"secret"`
}

// SetUpConfig sets up the correct configuration for the app
func SetUpConfig() (cfg Config, err error) {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	err = envconfig.Process("", &cfg)
	if err != nil {
		return
	}

	fmt.Println(cfg.DatabaseURL)
	return cfg, err
}
