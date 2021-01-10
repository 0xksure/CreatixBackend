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
	"github.com/kristohberg/CreatixBackend/internal/mail"
	"github.com/kristohberg/CreatixBackend/logging"
	jwtmiddleware "github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	db, err := sql.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return db, nil
}

func (a *App) MigrateUpDatabase(cfg config.Config) error {
	m, err := migrate.New("file://db/migrations", cfg.DatabaseUrl)
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
		return a, errors.Wrap(err, "not able to migrate up database")
	}
	a.DB = db
	a.cfg = cfg

	return a, nil
}

// Run starts up the application
func (a App) Run() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3030", "https://thecreatix.herokuapp.com", "https://thecreatix.io"},
		AllowHeaders:     []string{"authorization", "Content-Type"},
		AllowCredentials: true,
		AllowMethods:     []string{echo.OPTIONS, echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	port := fmt.Sprintf(":%s", a.cfg.Port)
	if port == "" {
		port = ":8000"
	}
	sessionClient := models.NewSessionClient(a.DB, []byte(a.cfg.TokenSecret), a.cfg.TokenExpirationTimeMinutes, a.logger)
	openSubrouter := e.Group("/v0")
	restAPI := handler.RestAPI{
		DB:             a.DB,
		Logging:        a.logger,
		Cfg:            a.cfg,
		Feedback:       models.Feedback{},
		Middleware:     &jwtmiddleware.Middleware{Cfg: a.cfg},
		CompanyClient:  models.NewCompanyClient(a.DB),
		SessionClient:  sessionClient,
		FeedbackClient: models.NewFeedbackClient(a.DB),
	}
	restAPI.Handler(openSubrouter)

	authSubrouter := e.Group("/v0/auth")
	sessionAPI := handler.SessionAPI{
		DB:            a.DB,
		Logging:       a.logger,
		Cfg:           a.cfg,
		SessionClient: sessionClient,
	}
	sessionAPI.Handler(authSubrouter)

	publicSubrouter := e.Group("/v0/public")
	publicAPI := handler.PublicAPI{
		Cfg:        a.cfg,
		Logging:    logrus.StandardLogger(),
		MailClient: mail.NewMailClient(a.cfg.SendgridKey),
	}
	publicAPI.Handler(publicSubrouter)

	// REST API handler
	log.Fatal(e.Start(port))
}
