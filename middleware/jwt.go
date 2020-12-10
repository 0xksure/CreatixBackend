package middleware

import (
	"net/http"
	"sync"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/labstack/echo"
)

type Exception struct {
	Message string `json:"Message"`
}

type Middleware struct {
	Claim *MiddlewareClaim
	Uid   string
	mutex sync.RWMutex
	Cfg   config.Config
}

type MiddlewareClaim struct {
	jwt.StandardClaims
}

func (m Middleware) JwtVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		m.mutex.Lock()
		defer m.mutex.Unlock()
		cookie, err := c.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				return c.JSON(http.StatusUnauthorized, "unauthorized: missing token")
			}
			return c.JSON(http.StatusBadRequest, Exception{Message: "bad request"})
		}

		tokenValue := cookie.Value
		err = utils.IsTokenValid(tokenValue, []byte(m.Cfg.TokenSecret))
		if err != nil {
			return c.JSON(http.StatusBadRequest, Exception{Message: err.Error()})
		}
		claims, err := utils.GetClaims(tokenValue, []byte(m.Cfg.TokenSecret))
		if err != nil {
			return c.JSON(http.StatusBadRequest, Exception{Message: err.Error()})
		}
		expiresAt := time.Now().Add(time.Minute * time.Duration(m.Cfg.TokenExpirationTimeMinutes))
		cookie.Expires = expiresAt
		c.SetCookie(cookie)
		c.Set(utils.UserIDContext.String(), claims.UserID)

		return next(c)
	}
}
