package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

const ioTimeout = time.Second * 3
const cacheWriteTimeout = time.Second * 30

type App struct {
	cfg    *config.Config
	echo   *echo.Echo
	DB     *sql.DB
	logger *logging.StandardLogger
}

type (
	RequestHandlerFunction func(a *App, w http.ResponseWriter, r *http.Request)
	HandlerFunc            func(w http.ResponseWriter, r *http.Request)
)

func ConnectDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DbURI)
	if err != nil {
		return db, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return db, nil
}

func (a *App) MigrateUpDatabase(cfg *config.Config) error {
	m, err := migrate.New("file://db/migrations", a.cfg.DbURI)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil {
		return err
	}
	return nil
}

// New sets up a new app
func New(cfg *config.Config) (App, error) {
	var a App
	a.logger = logging.NewLogger()

	db, err := ConnectDB(cfg)
	if err != nil {
		return a, errors.Wrap(err, "could not establish contact with the database")
	}

	err = db.Ping()
	if err != nil {
		return a, errors.Wrap(err, "could not ping database")
	}
	err = a.MigrateUpDatabase(a.cfg)
	if err != nil {
		return a, errors.Wrap(err, "not able to migrate up database")
	}
	a.DB = db

	// set up echo
	e := echo.New()
	e.Server.WriteTimeout = ioTimeout
	e.Server.ReadTimeout = ioTimeout
	e.HideBanner = true
	e.HidePort = true

	// Set up global middleware
	e.Use(middleware.Recover(), middleware.CSRF())

	a.echo = e
	a.cfg = cfg

	return a, nil
}

// Run starts up the application
func (a *App) Run() {
	port := a.cfg.DbPort
	if port == "" {
		port = "8000"
	}

	authSubrouter := a.echo.Group("/v0/auth", middleware.CORS())
	sessionAPI := handler.Session{
		DB:      a.DB,
		Logging: a.logger,
		Cfg:     a.cfg,
		UserSession: models.UserSession{
			JwtSecret: a.cfg.JwtSecret,
		},
	}
	sessionAPI.Handler(authSubrouter)

	openSubrouter := a.echo.Group("/v0/", middleware.CORS())
	restAPI := handler.RestAPI{
		DB:       a.DB,
		Logging:  a.logger,
		Cfg:      a.cfg,
		Feedback: models.Feedback{},
	}

	restAPI.Handler(openSubrouter)
	// REST API handler
	log.Fatal(a.echo.Start(port))
}
