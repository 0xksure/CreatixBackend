package utils

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type contextKey string

const UserIDContext = contextKey("userID")

func (u contextKey) String() string {
	return string(u)
}

func GetUserIDFromContext(c echo.Context) (string, error) {
	userID := c.Get(UserIDContext.String())
	if userID == nil {
		return "", errors.New("userID not passed along")
	}
	return userID.(string), nil
}
