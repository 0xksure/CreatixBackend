package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
)

type RestAPI struct {
	api     RestAPIModels
	notAuth []string
}

type RestAPIModels interface {
	ValidatAccounts(*models.Account) (map[string]interface{}, bool)
	ValidateContact(*models.Contact) (map[string]interface{}, bool)
	Create(*models.Account, http.ResponseWriter) map[string]interface{}
	Login(string, string) map[string]interface{}
	GetUser(uint) *models.Account
	GetDB() *gorm.DB
	GetToken(tokenHeader string) ([]string, map[string]interface{}, error)
}

func Restapi(cfg *config.Config, stdLog *logging.StandardLogger) {
	var restAPI RestAPI
	restAPI.notAuth = []string{"/api/v1/user/new", "/api/v1/user/login", "/api/v1/signup-newsletter", "/api/v1/newsletter/signup"}
	fmt.Println("OriginAllowed: ", cfg.OriginAllowed)
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Access-Control-Allow-Origin", "Content-Type"})
	originsOk := handlers.AllowedOrigins([]string{cfg.OriginAllowed})

	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	router := mux.NewRouter().StrictSlash(true)
	router.Use(restAPI.JwtAuthentication)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router.HandleFunc("/", homeLink)
	restAPI.Users(router)
	restAPI.NewsletterAPI(router)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to this awesome API")
}
