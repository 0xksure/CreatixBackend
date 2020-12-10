package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Claims struct {
	UserID string
	jwt.StandardClaims
}

// GetClaims takes a token string and the jwt secret and return a
// claim
func GetClaims(tokenString string, secret []byte) (*Claims, error) {
	claims := new(Claims)
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return claims, errors.Wrap(err, "signature invalid")
		}
		return claims, err
	}

	if !token.Valid {
		return claims, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return claims, errors.New("could not get claims")
	}
	return claims, nil
}

// IsTokenValid checks to see if a token is valid
func IsTokenValid(tokenString string, secret []byte) error {
	claims, err := GetClaims(tokenString, secret)
	if err != nil {
		return err
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 60*time.Second {
		return errors.New("invalid token")
	}

	return nil
}
