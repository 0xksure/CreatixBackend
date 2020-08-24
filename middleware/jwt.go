package jwtmiddleware

import (
	"fmt"
	"net/http"
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

type Exception struct {
	Message string `json:"Message"`
}

type Middleware struct {
	Claim *MiddlewareClaim
	Uid   string
	mutex sync.RWMutex
}

type MiddlewareClaim struct {
	jwt.StandardClaims
}

func (m *Middleware) JwtVerify(next echo.HandlerFunc) echo.HandlerFunc {
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

		tknStr := cookie.Value
		token, err := jwt.ParseWithClaims(tknStr, &MiddlewareClaim{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				return c.JSON(http.StatusUnauthorized, Exception{Message: "bad request"})
			}
			return c.JSON(http.StatusUnauthorized, Exception{Message: fmt.Sprintf("err1: %s", err.Error())})
		}

		claims, ok := token.Claims.(*MiddlewareClaim)
		if !ok || !token.Valid {
			return c.JSON(http.StatusUnauthorized, Exception{Message: fmt.Sprintf("err2: %s", err.Error())})
		}
		m.Uid = claims.Id
		fmt.Println("claims: ", claims)
		fmt.Println("middleware: ", m)
		fmt.Println("uid: ", claims.Id)
		return next(c)
	}
}
