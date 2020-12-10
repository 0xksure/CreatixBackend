package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/utils"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//

type SessionClient interface {
	CreateUser(ctx context.Context, signup Signup) error
	LoginUser(ctx context.Context, loginRequest *LoginRequest) (Response, error)
}

type sessionClient struct {
	DB                  *sql.DB
	TokenSecret         []byte
	TokenExpirationTime int
	logger              *logging.StandardLogger
}

// NewSessionClient creates new session client
func NewSessionClient(DB *sql.DB, tokenSecret []byte, tokenExpirationTime int, logger *logging.StandardLogger) sessionClient {
	return sessionClient{DB: DB, TokenSecret: tokenSecret, TokenExpirationTime: tokenExpirationTime, logger: logger}
}

// LoginRequest contains the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SessionUser struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}
type User struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}
type Signup struct {
	User
	Company
}

type FieldErrors map[string]string

func (fe FieldErrors) Error() string {
	var fieldErrorString string
	for key, val := range fe {
		fieldErrorString += fmt.Sprintf("key: %s, val: %s", key, val)
	}
	return fieldErrorString

}

func (s Signup) Valid() error {

	errs := make(FieldErrors)
	if s.Company.Name == "" {
		errs["companyName"] = "company name cannot be empty"
	}

	if s.User.Firstname == "" {
		errs["firstName"] = "firstname cannot be empty"
	}

	if s.User.Lastname == "" {
		errs["lastName"] = "lastname cannot be empty"
	}

	if s.User.Email == "" {
		errs["email"] = "email cannot be empty"
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type Response struct {
	Status      bool        `json:"status"`
	Message     string      `json:"message"`
	Token       string      `json:"token"`
	ExpiresAt   time.Time   `json:"expiresAt"`
	SessionUser SessionUser `json:"sessionUser"`
}

// NewToken creates a new token with a default claim
func (c sessionClient) newToken(expiresAt time.Time, userID string) (string, error) {

	claims := utils.Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Issuer:    "creatix",
			Id:        userID,
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, err := token.SignedString(c.TokenSecret)
	if err != nil {
		c.logger.Unsuccessful("not able to generate token string", err)
		return tokenString, err
	}
	return tokenString, nil

}

// LoginUser checks if the user given password and username exists
// if it does
func (c sessionClient) LoginUser(ctx context.Context, loginRequest *LoginRequest) (resp Response, err error) {
	existingUser, err := c.findUserByEmail(ctx, loginRequest.Email)
	if err != nil {
		c.logger.Unsuccessful("could not find user", err)
		return
	}

	hashedPassword, err := c.findUserPasswordByEmail(ctx, loginRequest.Email)
	if err != nil {
		c.logger.Unsuccessful("could not find user", err)
		return
	}

	errf := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginRequest.Password))
	if errf == bcrypt.ErrMismatchedHashAndPassword {
		c.logger.Unsuccessful("incorrect email or password", err)
		err = errors.New("passwords do not match")
		return
	}

	if err != nil {
		c.logger.Unsuccessful("incorrect email or password", err)
		return
	}

	expiresAt := time.Now().Local().Add(time.Minute * time.Duration(c.TokenExpirationTime))
	tokenString, err := c.newToken(expiresAt, existingUser.ID)
	if err != nil {
		c.logger.Unsuccessful("not able to generate token", err)
		return
	}

	resp.Status = false
	resp.Message = "logged in"
	resp.Token = tokenString
	resp.ExpiresAt = expiresAt
	resp.SessionUser = existingUser

	return resp, nil
}

var createUserCompanyQuery = `
WITH new_company AS (
	INSERT INTO company(Name)
	SELECT CAST($5 AS VARCHAR)
	WHERE NOT EXISTS (SELECT * FROM COMPANY WHERE Name=$5)
	RETURNING *
),
new_user AS (
	INSERT INTO users(firstname,lastname,email,password)
	VALUES ($1,$2,$3,$4)
	RETURNING *
)

INSERT INTO USER_COMPANY(CompanyId, UserId)
values ((SELECT Id from new_company),(SELECT ID FROM new_user))
`

var createUserQuery = `
	INSERT INTO users(firstname,lastname,email,password)
	VALUES ($1,$2,$3,$4)
`

// CreateUser creates a new user in the database
func (c sessionClient) CreateUser(ctx context.Context, signup Signup) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signup.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	res, err := c.DB.ExecContext(ctx, createUserQuery, signup.Firstname, signup.Lastname, signup.Email, hashedPassword)
	if err != nil {
		return err
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if nrows == 0 {
		return errors.New("not able to add user ")
	}

	return nil
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
func (c sessionClient) findUserByEmail(ctx context.Context, email string) (user SessionUser, err error) {
	err = c.DB.QueryRowContext(ctx, findUserByEmailQuery, email).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email)
	if err != nil {
		return user, err
	}
	return user, nil
}

var findUserPasswordByEmailQuery = `
	SELECT 
	Password
	FROM users
	WHERE Email = $1
`

func (c sessionClient) findUserPasswordByEmail(ctx context.Context, email string) (password string, err error) {
	err = c.DB.QueryRowContext(ctx, findUserPasswordByEmailQuery, email).Scan(&password)
	if err != nil {
		return
	}
	return
}
