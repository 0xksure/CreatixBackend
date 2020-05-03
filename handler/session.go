package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/utils"

	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/models"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	DB          *sql.DB
	Logging     *logging.StandardLogger
	Cfg         *config.Config
	UserSession models.UserSession
}

// LoginRequest contains the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handler sets up the session endpoints
func (s Session) Handler(e *echo.Group) {
	e.POST("/user/signup", s.Signup)
	e.POST("/user/login", s.Login)
	e.POST("/user/refresh", s.Refresh)
	e.GET("/user/logout", s.Logout)
}

// Signup signups the new user
func (s Session) Signup(c echo.Context) error {
	var user models.User

	err := c.Bind(user)
	if err != nil {
		s.Logging.Unsuccessful("could not parse user data", err)
		return c.String(http.StatusBadRequest, "could not bind data")
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logging.Unsuccessful("not able to encrypt password", err)
		return c.String(http.StatusBadRequest, "could not generate hashed password")
	}

	user.Password = string(pass)
	err = s.UserSession.CreateUser(c.Request().Context(), s.DB, user)
	if err != nil {
		s.Logging.Unsuccessful("not able to create user ", err)
		return c.String(http.StatusBadRequest, "could not create user")
	}

	return c.String(http.StatusOK, "user created")
}

// Login checks whether the user exists and creates a cookie
func (s Session) Login(c echo.Context) error {
	var loginRequest LoginRequest
	err := c.Bind(loginRequest)
	if err != nil {
		s.Logging.Unsuccessful("not able to parse user", err)
		return c.String(http.StatusBadRequest, "not able to parse user")
	}
	resp, err := s.UserSession.LoginUser(c.Request().Context(), s.DB, loginRequest.Email, loginRequest.Password)
	if err != nil {
		s.Logging.Unsuccessful("not able to log in user", err)
		return c.String(http.StatusBadRequest, "not able to parse user")
	}
	cookie := &http.Cookie{
		Name:    "token",
		Value:   resp.Token,
		Expires: resp.ExpiresAt,
		Path:    "/v0",
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, resp)
}

// Logout will set a new invalid cookie
func (s Session) Logout(c echo.Context) error {
	cookie := &http.Cookie{
		Name:   "token",
		MaxAge: -1,
		Path:   "/v0",
	}
	c.SetCookie(cookie)
	return c.String(http.StatusOK, "old cookie deleted, logged out")
}

// ForgotPassword will if the user exists send a
// new passwordlink
func (s Session) ForgotPassword(c echo.Context) error {
	var forgotpassword models.PasswordRequest
	err := c.Bind(forgotpassword)
	if err != nil {
		s.Logging.Unsuccessful("not able to parse user", err)
		return c.String(http.StatusBadRequest, "not able to parse email")
	}

	resp, err := s.UserSession.ForgotPassword(c, s.DB, forgotpassword.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, "not able to generate new password")
	}

	return c.JSON(http.StatusOK, resp)

}

// Refresh refreshes the cookie provided by generating a new one
// and returning it
func (s Session) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("token")
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "invalid cookie"})
	}

	tokenValue := cookie.Value
	ok, err := utils.IsTokenValid(tokenValue, []byte(s.Cfg.JwtSecret))
	if err != nil || !ok {
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "invalid cookie"})
	}

	expiresAt := time.Now().Add(time.Minute * 5)
	newToken, err := utils.NewToken(expiresAt, "1", []byte(s.Cfg.JwtSecret))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "could not generate new token"})
	}
	c.SetCookie(&http.Cookie{
		Name:    "token",
		Value:   newToken,
		Expires: expiresAt,
		Path:    "/v0",
	})
	return c.JSON(http.StatusOK, utils.HttpResponse{Message: "ok"})
}
