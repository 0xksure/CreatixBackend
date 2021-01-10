package utils

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var (
	NoPermission = errors.New("user does not have permission")
)

type SessionUser struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}

var findUserByEmailQuery = `
	SELECT 
	ID
	,Firstname
	,Lastname
	,Email
	FROM users
	WHERE Email = $1
`

// findUserByEmail returns the first row with the given email
func FindUserByEmail(ctx context.Context, DB *sql.DB, email string) (user SessionUser, err error) {
	err = DB.QueryRowContext(ctx, findUserByEmailQuery, email).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email)
	if err != nil {
		return user, errors.WithMessagef(err, "feedback.utils.finduserbyemail")
	}
	return user, nil
}
