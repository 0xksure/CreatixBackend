package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type JwtSession struct {
	Secret    []byte
	ExpiryMin int
}

type Claims struct {
	UserID string
	jwt.StandardClaims
}

// NewToken creates a new token with a default claim
func NewToken(expiresAt time.Time, userID string, secret []byte) (string, error) {

	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "creatix",
			Id:        userID,
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return tokenString, err
	}
	return tokenString, nil

}

// GetClaims takes a token string and the jwt secret and return a
// claim
func GetClaims(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.Wrap(err, "signature invalid")
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("could not get claims")
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
