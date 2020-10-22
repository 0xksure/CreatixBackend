package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env               string `default:""`
	DbUser            string `default:""`
	DbPass            string `default:""`
	DbName            string `default:""`
	DbHost            string `default:""`
	DbPort            string `default:""`
	DbURI             string `default:""`
	DatabaseUrl       string `split_words:"true"`
	Port              string `split_words:"true" default:":8080"`
	OriginAllowed     string `default:""`
	FromEmail         string `default:""`
	SMTPServer        string `default:""`
	SMTPPWD           string `default:""`
	SMTPUserName      string `default:""`
	JwtSecret         string `default:"secret"`
	AllowCookieDomain string `split_words:"true" default:""`
}

// SetUpConfig sets up the correct configuration for the app
func SetUpConfig() (cfg Config, err error) {

	logrus.SetFormatter(&logrus.JSONFormatter{})
	err = envconfig.Process("", &cfg)
	if err != nil {
		return
	}
	return cfg, err
}
