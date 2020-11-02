package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/kristohberg/CreatixBackend/web"

	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/models"
	"golang.org/x/crypto/bcrypt"
)

type SessionAPI struct {
	DB          *sql.DB
	Logging     *logging.StandardLogger
	Cfg         config.Config
	UserSession models.UserSession
}

// LoginRequest contains the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handler sets up the session endpoints
func (s SessionAPI) Handler(e *echo.Group) {
	e.POST("/user/signup", s.Signup)
	e.POST("/user/login", s.Login)
	e.POST("/user/refresh", s.Refresh)
	e.GET("/user/logout", s.Logout)
}

// Signup signups the new user
func (s SessionAPI) Signup(c echo.Context) (err error) {
	signup := new(models.Signup)
	err = c.Bind(signup)
	if err != nil {
		s.Logging.Unsuccessful("could not parse user data", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "could not bind data"})
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(signup.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logging.Unsuccessful("not able to encrypt password", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "could not generate hashed password"})
	}

	signup.Password = string(pass)
	err = s.UserSession.CreateUser(c.Request().Context(), s.DB, *signup)
	if err != nil {
		s.Logging.Unsuccessful("not able to create user ", err)
		return c.String(http.StatusBadRequest, "could not create user")
	}

	return c.String(http.StatusOK, "user created")
}

// Login checks whether the user exists and creates a cookie
func (s SessionAPI) Login(c echo.Context) error {
	loginRequest := new(LoginRequest)
	err := c.Bind(loginRequest)
	if err != nil {
		s.Logging.Unsuccessful("not able to parse user", err)
		return c.String(http.StatusBadRequest, "not able to parse user")
	}

	resp, err := s.UserSession.LoginUser(c.Request().Context(), s.DB, loginRequest.Email, loginRequest.Password, s.Cfg.TokenExpirationTimeMinutes)
	if err != nil {
		s.Logging.Unsuccessful("not able to log in user", err)
		return c.String(http.StatusBadRequest, "no user")
	}
	cookie := &http.Cookie{
		Name:    "token",
		Value:   resp.Token,
		Expires: resp.ExpiresAt,
		Path:    "/v0",
		Domain:  s.Cfg.AllowCookieDomain,
	}
	if s.Cfg.Env == "prod" {
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = true
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, resp)
}

// Logout will set a new invalid cookie
func (s SessionAPI) Logout(c echo.Context) error {
	cookie := &http.Cookie{
		Name:   "token",
		MaxAge: -1,
		Path:   "/v0",
	}
	if s.Cfg.Env == "prod" {
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = true
	}
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "old cookie deleted, logged out")
}

// Refresh refreshes the cookie provided by generating a new one
// and returning it
func (s SessionAPI) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("token")
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "cookie not found"})
	}

	tokenValue := cookie.Value
	err = utils.IsTokenValid(tokenValue, []byte(s.Cfg.TokenSecret))
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: fmt.Sprintf("invalid cookie: %s", err.Error())})
	}

	expiresAt := time.Now().Add(time.Minute * time.Duration(s.Cfg.TokenExpirationTimeMinutes))
	cookie.Expires = expiresAt
	cookie.Path = "/v0"
	if s.Cfg.Env == "prod" {
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = true
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, "")
}
