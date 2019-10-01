package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/models"
	"github.com/kristofhb/CreatixBackend/utils"
)

type App struct {
	cfg    *config.Config
	router *mux.Router
	src    *http.Server
}

// New sets up a new app
func New(cfg *config.Config) {

}

var JwtAuthentication = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// endpoints that does not require auth
		notAuth := []string{"/api/user/new", "/api/user/login", "/"}

		// Current request path
		requestPath := r.URL.Path

		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string]interface{})
		// Grab token from header
		tokenHeader := r.Header.Get("Authorization")

		// token header is missing
		if tokenHeader == "" {
			response = utils.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		splitted := strings.Split(tokenHeader, " ")
		// Check if token comes in format `Bearer {token}`
		if len(splitted) != 2 {
			response = utils.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content Type", "application/json")
			utils.Respond(w, response)
			return
		}

		// Take bearer token
		tokenPart := splitted[1]
		tk := &models.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})
		if err != nil {
			response = utils.Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		// if token is not valid may not be signed by server
		if !token.Valid {
			response = utils.Message(false, "Token is invalid")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			utils.Respond(w, response)
			return
		}

		// Proceed with request and set the caller to the user retrieved from the parsed token
		fmt.Printf("user % ", tk.UserId)
		ctx := context.WithValue(r.Context(), "user", tk.UserId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
