package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Firstname string
	Lastname  string
	Birthday  time.Time
	Email     string `gorm:"type:varchar(100);unique_index"`
	Password  string `json:"password"`
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
func (user *User) CreateUser(ctx context.Context, db *sql.DB) error {
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
	Firstname
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
func (user *User) LoginUser(ctx context.Context, db *sql.DB, userEmail string) (Response, error) {

	var resp Response
	u, err := findUserByEmail(ctx, db, user.Email)
	if err != nil {
		return resp, err
	}

	errf := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(user.Password))
	if errf == bcrypt.ErrMismatchedHashAndPassword {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, errors.New("passwords do not match")
	}

	expiresAt := time.Now().Add(time.Minute * 30)
	us := UserSession{
		UserID:    authUser.ID,
		Firstname: authUser.Firstname,
		Lastname:  authUser.Lastname,
		Email:     authUser.Email,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), us)

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, errors.New("not able to sign string")
	}

	resp.Status = false
	resp.Message = "logged in"
	resp.Token = tokenString
	resp.ExpiresAt = expiresAt
	resp.User = authUser
	user.Firstname = authUser.Firstname
	user.Lastname = authUser.Lastname

	return resp, nil

}

func (pr *PasswordRequest) ForgotPassword(db *gorm.DB) (resp Response, err error) {
	// Create New password request
	var user User
	if err = db.Where("Email = ?", pr.Email).First(&user).Error; err != nil {
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
	hashedGuid, err := bcrypt.GenerateFromPassword([]byte(guid.String()), bcrypt.DefaultCost)
	if err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}

	pce := &PasswordChangeRequest{
		ReqID:  string(hashedGuid),
		UserID: user.ID,
	}

	// send mail to user
	net.smp
}
