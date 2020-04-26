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
	UserID uint
	jwt.StandardClaims
}

func NewJwtSession(secret []byte, expiryMin int) *JwtSession {
	return &JwtSession{
		Secret:    secret,
		ExpiryMin: expiryMin,
	}
}

// NewToken creates a new token with a default claim
func (js *JwtSession) NewToken(expiresAt time.Time, userID uint, secret []byte) (string, error) {

	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "creatix",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(secret)
	if err != nil {
		return ss, err
	}

	return ss, nil

}

// VerifyToken parses the token using the given secret to check if the token is valid
func (c Claims) VerifyToken(tokenString string, secret []byte) error {
	tkn, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return errors.Wrap(err, "user is not authorized: token could not be parsed")
	}

	if !tkn.Valid {
		return errors.New("user is not authorized: token is not valid")
	}

	return nil
}

func (c Claims) RefreshToken(tokenString string, secret []byte) error {

	err := c.VerifyToken(tokenString, secret)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Minute * 3)
}
