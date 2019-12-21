package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kristofhb/CreatixBackend/config"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	DB      *gorm.DB
	Logging *logging.StandardLogger
	cfg     *config.Config
}

func (s Session) Handler(r *mux.Router) {
	r.HandleFunc("/user/signup", s.Signup).Methods("POST")
	r.HandleFunc("/user/login", s.Login).Methods("POST")
	r.HandleFunc("/user/refresh-token", s.RefreshToken).Method("GET")
	r.HandleFunc("/user/logout", s.Logout).Method("GET")
}

func (s Session) Signup(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logging.Unsuccessful("not able to encrypt password", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = string(pass)
	createdUser, err := models.CreateUser(s.DB, *user)
	if err != nil {
		s.Logging.Unsuccessful("not able to create user ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err = json.NewEncoder(w).Encode(createdUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		s.Logging.Unsuccessful("not able to write out created user", err)
		return
	}
}

func (s Session) Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		s.Logging.Unsuccessful("not able to parse user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp models.Response
	resp, err = models.LoginUser(s.DB, *user)
	if err != nil {
		s.Logging.Unsuccessful("not able to log in user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}
