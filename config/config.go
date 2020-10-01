package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env           string `default:""`
	DbUser        string `default:""`
	DbPass        string `default:""`
	DbName        string `default:""`
	DbHost        string `default:""`
	DbPort        string `default:""`
	DbURI         string `default:""`
	ListenPort    string `default:":8080"`
	OriginAllowed string `default:""`
	FromEmail     string `default:""`
	SMTPServer    string `default:""`
	SMTPPWD       string `default:""`
	SMTPUserName  string `default:""`
	JwtSecret     string `default:"secret"`
}

// SetUpConfig sets up the correct configuration for the app
func SetUpConfig() (cfg Config, err error) {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	err = envconfig.Process("", &cfg)
	if err != nil {
		return
	}

	cfg.DbURI = fmt.Sprintf("postgres://%s:%s/%s?user=%s&password=%s&sslmode=disable", cfg.DbHost, cfg.DbPort, cfg.DbName, cfg.DbUser, cfg.DbPass)
	return cfg, err
}
