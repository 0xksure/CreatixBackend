package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kristohberg/CreatixBackend/utils"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//
var SigningKey = []byte("secret")

type User struct {
	ID        string    `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Birthday  time.Time `json:"birthday"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

type UserSession struct {
	JwtSecret string
	ID        string
}

func (u UserSession) Valid() error {
	return nil
}

type PasswordRequest struct {
	Email string
}

type PasswordChangeRequest struct {
	ReqID  string
	UserID string
}

type UserInformation struct {
	UserID      uint
	User        User
	PhoneNumber string
	BirthDate   string
	Gender      string
}

type Response struct {
	Status    bool
	Message   string
	Token     string
	ExpiresAt time.Time
	User
}

var cookieExpireTime = 30 * time.Minute

var createUserQuery = `
	INSERT INTO users(firstname,lastname,birthday,email,password)
	VALUES ($1,$2,$3,$4,$5)
`

// CreateUser creates a new user in the database
func (u UserSession) CreateUser(ctx context.Context, db *sql.DB, user User) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	res, err := tx.Exec(createUserQuery, user.Firstname, user.Lastname, user.Birthday, user.Email, user.Password)
	if err != nil {
		fmt.Println("err: ", err)
		err = tx.Rollback()
		if err != nil {
			return err
		}
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errors.New("0 rows affected")
	}
	fmt.Println("err: ", err)

	return nil
}

var findUserByEmailQuery = `
	SELECT 
	ID
	,Firstname
	,Lastname
	,Birthday
	,Email
	,Password
	FROM users
	WHERE Email = $1
`

// findUserByEmail returns the first row with the given email
func findUserByEmail(ctx context.Context, db *sql.DB, email string) (User, error) {
	var user User
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return user, err
	}

	err = tx.QueryRow(findUserByEmailQuery, email).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Birthday, &user.Email, &user.Password)
	if err != nil {
		return user, err
	}

	return user, nil
}

// LoginUser checks if the user given password and username exists
// if it does
func (u *UserSession) LoginUser(ctx context.Context, db *sql.DB, userEmail string, password string) (Response, error) {
	var resp Response
	existingUser, err := findUserByEmail(ctx, db, userEmail)
	if err != nil {
		return resp, err
	}

	errf := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password))
	if errf == bcrypt.ErrMismatchedHashAndPassword {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, errors.New("passwords do not match")
	}

	expiresAt := time.Now().Add(time.Minute * 30)
	tokenString, err := utils.NewToken(expiresAt, existingUser.ID, []byte("secret"))
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}

	resp.Status = false
	resp.Message = "logged in"
	resp.Token = tokenString
	resp.ExpiresAt = expiresAt

	u.ID = existingUser.ID

	return resp, nil
}

// ForgotPassword send a new password link
/*
func (u UserSession) ForgotPassword(ctx context.Context, db *sql.DB, email string) (resp Response, err error) {
	// Create New password request
	user, err := findUserByEmail(ctx, db, email)
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}

	// Create request ID
	guid, err := uuid.NewRandom()
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}

	// Has gui
	hashedGUID, err := bcrypt.GenerateFromPassword([]byte(guid.String()), bcrypt.DefaultCost)
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}

	pce := &PasswordChangeRequest{
		ReqID:  string(hashedGUID),
		UserID: user.ID,
	}

	// send mail to user
	return resp, nil
}
*/
