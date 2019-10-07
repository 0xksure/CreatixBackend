package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/models"
	"github.com/kristofhb/CreatixBackend/utils"
)

func (a *RestAPI) Users(r *mux.Router) {
	r.HandleFunc("/api/v1/user/new", a.CreateAccount).Methods("POST")
	r.HandleFunc("/api/v1/user/login", a.Authenticate).Methods("POST")
}

func (a *RestAPI) CreateAccount(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	// Decode request body into Account struct
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		utils.Respond(w, utils.Message(false, "Invalid request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create account
	resp := models.Create(account, w)
	utils.Respond(w, resp)

}

func (a *RestAPI) Authenticate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Authenitcate'")
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)

	if err != nil {
		utils.Respond(w, utils.Message(false, "Invalid request"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println("Login")
	resp := models.Login(account.Email, account.Password)
	utils.Respond(w, resp)

}

func (a RestAPI) JwtAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// endpoints that does not require auth

		// Current request path
		requestPath := r.URL.Path

		for _, value := range a.notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		response := make(map[string]interface{})
		// Grab token from header
		tokenHeader := r.Header.Get("Authorization")
		fmt.Println("Token header")
		fmt.Println(tokenHeader)
		splitted, msg, err := models.GetToken(tokenHeader)
		fmt.Println("OK")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content type", "application/json")
			utils.Respond(w, msg)
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
		fmt.Printf("user % ", tk.UserID)
		ctx := context.WithValue(r.Context(), "user", tk.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
