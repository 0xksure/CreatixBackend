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

func NewJwtSession(secret []byte, expiryMin int) *JwtSession {
	return &JwtSession{
		Secret:    secret,
		ExpiryMin: expiryMin,
	}
}

// NewToken creates a new token with a default claim
func NewToken(expiresAt time.Time, userID string, secret []byte) (string, error) {

	claims := Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "creatix",
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return tokenString, err
	}

	return tokenString, nil

}

// VerifyToken parses the token using the given secret to check if the token is valid
func IsTokenValid(tokenString string, secret []byte) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return false, errors.Wrap(err, "signature invalid")
		}
		return false, err
	}

	if !token.Valid {
		return false, errors.New("toekn is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false, errors.New("could not get claims")
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 2*time.Minute {
		return false, errors.New("invalid token")
	}

	return true, nil
}

func RefreshToken(tokenString string, secret []byte) error {

	err := c.VerifyToken(tokenString, secret)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Minute * 3)
}
