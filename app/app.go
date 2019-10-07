package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/config"
)

type App struct {
	cfg    *config.Config
	router *mux.Router
	src    *http.Server
}

// New sets up a new app
func New(cfg *config.Config) {

}
