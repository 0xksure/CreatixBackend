package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kristofhb/CreatixBackend/api"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
)

type App struct {
	cfg    *config.Config
	Router *mux.Router
	src    *http.Server
	db     *gorm.DB
	logger *logging.StandardLogger
	api    api.Feedback
}

type (
	RequestHandlerFunction func(a *App, w http.ResponseWriter, r *http.Request)
	HandlerFunc            func(w http.ResponseWriter, r *http.Request)
)

// New sets up a new app
func (a *App) New(cfg *config.Config) {
	a.logger = logging.NewLogger()
	a.api = api.Feedback
	dbUri := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbName, cfg.DbPass)

	var err error
	a.db, err = gorm.Open("postgres", dbUri)
	if err != nil {
		a.logger.Unsuccessful("not able to open postgres", err)
		log.Fatal(err)
	}

	a.db.Debug().AutoMigrate(&models.User)

	a.Router = mux.NewRouter().StrictSlash(true)

	a.cfg = cfg
}

func (a *App) Run() {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Access-Control-Allow-Origin", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{a.cfg.OriginAllowed})

	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	port := a.cfg.DbPort
	if port == "" {
		port = "8000"
	}

	sessionAPI := handler.Session{
		DB:      a.db,
		Logging: a.logger,
		cfg: a.cfg
	}
	sessionAPI.Handler(a.Router)
	// REST API handler
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk)(a.Router)))
}
