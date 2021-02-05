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

var findUserByUserIdQuery = `
	SELECT 
	ID
	,Firstname
	,Lastname
	,Email
	FROM users
	WHERE ID = $1
`

func FindUserByUserID(ctx context.Context, DB *sql.DB, userID string) (user SessionUser, err error) {
	err = DB.QueryRowContext(ctx, findUserByUserIdQuery, userID).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email)
	if err != nil {
		return user, errors.WithMessagef(err, "feedback.utils.finduserbyuserid")
	}
	return user, nil
}

var findUserByUsernameQuery = `
	SELECT 
	ID
	,Firstname
	,Lastname
	,Email
	FROM users
	WHERE username = $1
`

func FindUserByUsername(ctx context.Context, DB *sql.DB, username string) (user SessionUser, err error) {
	err = DB.QueryRowContext(ctx, findUserByUsernameQuery, username).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email)
	if err != nil {
		return user, errors.WithMessagef(err, "feedback.utils.finduserbyuserid")
	}
	return user, nil
}
