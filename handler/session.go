package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

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
	Cfg     *config.Config
	User    *models.User
}

type HttpResponse struct {
	Message string
	User    models.UserSession
}

func (s Session) Handler(r *mux.Router) {
	r.HandleFunc("/user/signup", s.Signup).Methods("POST")
	r.HandleFunc("/user/login", s.Login).Methods("POST")
	r.HandleFunc("/user/logout", s.Logout).Methods("GET")
	r.HandleFunc("/user/refresh", s.Refresh).Methods("POST")

}

func (s Session) Signup(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
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
	createdUser, err := user.CreateUser(s.DB)
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
	user := s.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		s.Logging.Unsuccessful("not able to parse user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp models.Response
	resp, err = user.LoginUser(s.DB)
	if err != nil {
		s.Logging.Unsuccessful("not able to log in user", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cookie := &http.Cookie{
		Name:    "token",
		Value:   resp.Token,
		Expires: resp.ExpiresAt,
		Path:    "/v0",
	}
	http.SetCookie(w, cookie)
	json.NewEncoder(w).Encode(resp)
}

func (s Session) Logout(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:   "token",
		MaxAge: -1,
		Path:   "/v0",
	}
	http.SetCookie(w, &c)
	w.Write([]byte("Old cookie deleted. Logget out!\n"))
}

func (s Session) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			s.Logging.Unsuccessful("not cookie set", err)
			w.Write([]byte("no cookie set"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(HttpResponse{Message: "not able to find cookie"})
		s.Logging.Unsuccessful("not able to find cookie", err)
		return
	}

	tknStr := c.Value
	us := &models.UserSession{}
	tkn, err := jwt.ParseWithClaims(tknStr, us, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(HttpResponse{Message: "not able to parse"})
			s.Logging.Unsuccessful("not able to parse", err)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(HttpResponse{Message: "bad request"})
		s.Logging.Unsuccessful("bad request", err)
		return
	}

	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(HttpResponse{Message: "token not valid"})
		s.Logging.Unsuccessful("token not valid", err)
		return
	}

	if time.Unix(us.ExpiresAt, 0).Sub(time.Now()) > 2*time.Minute {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HttpResponse{Message: "previous token has not expired", User: *us})
		s.Logging.Unsuccessful("previous token has not expired", err)
		return
	}

	expiresAt := time.Now().Add(time.Minute * 5)
	us.ExpiresAt = expiresAt.Unix()
	// Create new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, us)
	tokenString, err := token.SignedString(tkn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HttpResponse{Message: "not able to generate token string"})
		s.Logging.Unsuccessful("creatix.auth.refresh, not able to generate token string", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expiresAt,
		Path:    "/v0",
	})

}
