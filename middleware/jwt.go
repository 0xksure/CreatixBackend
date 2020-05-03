package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kristofhb/CreatixBackend/utils"
	"github.com/labstack/echo"
)

type Exception struct {
	Message string `json:"Message"`
}

func JwtVerify(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil{
			return c.JSON(http.StatusInternalServerError,utils.HttpResponse{Message:"could not serve http requests"})
		}
		cookie, err := c.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				return c.JSON(http.StatusUnauthorized,"unauthorized: missing token")
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Exception{Message: "bad request"})
			return c.JSON(http.StatusBadRequest,"bad request")
		}

		tk := &utils.UserSession{}
		tknStr := c.Value
		tkn, err := jwt.ParseWithClaims(tknStr, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(Exception{Message: err.Error()})
			return
		}

		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), "user", tk)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
