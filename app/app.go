package app

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/handler"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
)

type App struct {
	cfg      *config.Config
	Router   *mux.Router
	src      *http.Server
	db       *gorm.DB
	logger   *logging.StandardLogger
	user     *models.User
	feedback *models.Feedback
	smpt     *smtp.Client
}

type (
	RequestHandlerFunction func(a *App, w http.ResponseWriter, r *http.Request)
	HandlerFunc            func(w http.ResponseWriter, r *http.Request)
)

// New sets up a new app
func (a *App) New(cfg *config.Config) {
	a.logger = logging.NewLogger()
	dbUri := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbName, cfg.DbPass)

	var err error
	a.db, err = gorm.Open("postgres", dbUri)
	if err != nil {
		a.logger.Unsuccessful("not able to open postgres", err)
		log.Fatal(err)
	}
	a.db.LogMode(true)
	a.db.Debug().AutoMigrate(&models.User{})
	a.db.Debug().AutoMigrate(&models.Clap{})
	a.db.Debug().AutoMigrate(&models.Comment{})
	a.db.Debug().AutoMigrate(&models.Feedback{})

	// SMTP client
	host, _, _ := net.SplitHostPort(cfg.SMTPServer)
	auth := smtp.PlainAuth("", "username@example.tld", "password", host)

	r := mux.NewRouter().StrictSlash(true)
	//r.Use(CommonMiddleware)
	a.Router = r
	a.cfg = cfg
	a.user = &models.User{}
}

func (a *App) Run() {
	headersOk := handlers.AllowedHeaders([]string{"Accept", "content-type", "X-Requested-With", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Screen"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	credentialsOK := handlers.AllowCredentials()
	port := a.cfg.DbPort
	if port == "" {
		port = "8000"
	}

	router := a.Router
	sessionRouter := router.PathPrefix("/v0/auth").Subrouter()
	sessionAPI := handler.Session{
		DB:      a.db,
		Logging: a.logger,
		Cfg:     a.cfg,
		User:    a.user,
	}
	sessionAPI.Handler(sessionRouter)

	restRouter := router.PathPrefix("/v0").Subrouter()
	restAPI := handler.RestAPI{
		DB:       a.db,
		Logging:  a.logger,
		Cfg:      a.cfg,
		User:     a.user,
		Feedback: a.feedback,
	}

	restAPI.Handler(restRouter)
	// REST API handler
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, headersOk, methodsOk, credentialsOK)(a.Router)))
}

func CommonMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000/")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Header", "Accept,Content-Type, Content-Length,Accept-Encoding, X-CSRF-Token,Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}
