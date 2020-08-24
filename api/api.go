package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/handler"
	"github.com/kristohberg/CreatixBackend/logging"
	jwtmiddleware "github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const ioTimeout = time.Second * 3
const cacheWriteTimeout = time.Second * 30

type App struct {
	cfg    config.Config
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
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return db, nil
}

func (a *App) MigrateUpDatabase(cfg config.Config) error {
	m, err := migrate.New("file://db/migrations", cfg.DbURI)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// New sets up a new app
func New(cfg config.Config) (App, error) {
	var a App
	a.logger = logging.NewLogger()

	db, err := ConnectDB(&cfg)
	if err != nil {
		return a, errors.Wrap(err, "could not establish contact with the database")
	}

	err = db.Ping()
	if err != nil {
		return a, errors.Wrap(err, "could not ping database")
	}
	err = a.MigrateUpDatabase(cfg)
	if err != nil {
		fmt.Println("err: ", err)
		return a, errors.Wrap(err, "not able to migrate up database")
	}
	a.DB = db
	fmt.Println("db1: ", a.DB)
	a.cfg = cfg

	return a, nil
}

// Run starts up the application
func (a App) Run() {
	e := echo.New()
	e.Server.WriteTimeout = ioTimeout
	e.Server.ReadTimeout = ioTimeout
	e.HideBanner = true
	e.HidePort = true

	// Set up global middleware
	e.Use(middleware.Recover())

	port := a.cfg.ListenPort
	if port == "" {
		port = ":8000"
	}
	userSession := &models.UserSession{JwtSecret: a.cfg.JwtSecret}
	openSubrouter := e.Group("/v0")
	restAPI := handler.RestAPI{
		DB:          a.DB,
		Logging:     a.logger,
		Cfg:         a.cfg,
		Feedback:    models.Feedback{},
		UserSession: userSession,
		Middleware:  &jwtmiddleware.Middleware{},
	}
	restAPI.Handler(openSubrouter)

	authSubrouter := e.Group("/v0/auth")
	sessionAPI := handler.Session{
		DB:          a.DB,
		Logging:     a.logger,
		Cfg:         a.cfg,
		UserSession: *userSession,
	}
	sessionAPI.Handler(authSubrouter)

	// REST API handler
	log.Fatal(e.Start(port))
}