package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/kristofhb/CreatixBackend/utils"
	"golang.org/x/crypto/bcrypt"
)

//
var SigningKey = []byte("secret")

type User struct {
	ID        uint      `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Birthday  time.Time `json:"birthday"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
}

type UserSession struct {
	JwtSecret string
}

type PasswordRequest struct {
	Email string
}

type PasswordChangeRequest struct {
	gorm.Model
	ReqID  string
	UserID uint
}

type UserInformation struct {
	gorm.Model
	UserID      uint
	User        User `gorm:"foreignkey:UserID"`
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

var createUseQuery = `
	INSERT INTO user
	VALUES ($1,$2,$3,$4,$5)
`

// CreateUser creates a new user in the database
func (u UserSession) CreateUser(ctx context.Context, db *sql.DB, user User) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	_, err = tx.Exec(createUseQuery, user.Firstname, user.Lastname, user.Birthday, user.Email, user.Password)
	if err != nil {
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
	FROM user
	WHERE Email = $1
`

// findUserByEmail returns the first row with the given email
func findUserByEmail(ctx context.Context, db *sql.DB, email string) (User, error) {
	var user User
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return user, err
	}

	err = tx.QueryRow(findUserByEmailQuery, email).Scan(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// LoginUser checks if the user given password and username exists
// if it does
func (u UserSession) LoginUser(ctx context.Context, db *sql.DB, userEmail string, password string) (Response, error) {
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
	tokenString, err := utils.NewToken(expiresAt, existingUser.ID, []byte(u.JwtSecret))
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, errors.New("not able to sign string")
	}

	resp.Status = false
	resp.Message = "logged in"
	resp.Token = tokenString
	resp.ExpiresAt = expiresAt

	return resp, nil
}

// ForgotPassword send a new password link
func (u UserSession) ForgotPassword(ctx context.Context, db *sql.DB, email string) (resp Response, err error) {
	// Create New password request
	var user User
	user, err = findUserByEmail(ctx, db, email)
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
