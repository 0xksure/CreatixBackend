package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

//

type SessionClienter interface {
	CreateUser(ctx context.Context, signup Signup) error
	LoginUser(ctx context.Context, loginRequest *LoginRequest) (Response, error)
}

type SessionClient struct {
	DB                  *sql.DB
	TokenSecret         []byte
	TokenExpirationTime int
	logger              *logging.StandardLogger
}

// NewSessionClient creates new session client
func NewSessionClient(DB *sql.DB, tokenSecret []byte, tokenExpirationTime int, logger *logging.StandardLogger) *SessionClient {
	return &SessionClient{DB: DB, TokenSecret: tokenSecret, TokenExpirationTime: tokenExpirationTime, logger: logger}
}

// LoginRequest contains the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Signup struct {
	User
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

	if s.User.Firstname == "" {
		errs["firstName"] = "firstname cannot be empty"
	}

	if s.User.Lastname == "" {
		errs["lastName"] = "lastname cannot be empty"
	}

	if s.User.Username == "" {
		errs["username"] = "username cannot be empty"
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
	Status      bool              `json:"status"`
	Message     string            `json:"message"`
	Token       string            `json:"token"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	SessionUser utils.SessionUser `json:"sessionUser"`
}

// NewToken creates a new token with a default claim
func (c *SessionClient) newToken(expiresAt time.Time, userID string) (string, error) {

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
func (c *SessionClient) LoginUser(ctx context.Context, loginRequest *LoginRequest) (resp Response, err error) {
	existingUser, err := utils.FindUserByEmail(ctx, c.DB, loginRequest.Email)
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
	INSERT INTO users(firstname,lastname,username,email,password)
	VALUES ($1,$2,$3,$4,$5)
`

// CreateUser creates a new user in the database
func (c *SessionClient) CreateUser(ctx context.Context, signup Signup) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signup.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithMessage(err, "could not generate hashed password")
	}

	res, err := c.DB.ExecContext(ctx, createUserQuery, signup.Firstname, signup.Lastname, signup.Username, signup.Email, hashedPassword)
	if err != nil {
		return errors.WithMessage(err, "could not create user given data")
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return errors.WithMessage(err, "no rows affected")
	}

	if nrows == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

var findUserPasswordByEmailQuery = `
	SELECT 
	Password
	FROM users
	WHERE Email = $1
`

func (c *SessionClient) findUserPasswordByEmail(ctx context.Context, email string) (password string, err error) {
	err = c.DB.QueryRowContext(ctx, findUserPasswordByEmailQuery, email).Scan(&password)
	if err != nil {
		return
	}
	return
}

var isAuthorizedQuery = `
	SELECT 
	uc.UserId
	FROM USER_COMPANY as uc
	LEFT JOIN (
		SELECT 
		AccessID
		FROM COMPANY_ACCESS
	) as ca 
	ON ca.AccessID = uc.AccessId
	WHERE uc.CompanyId=$1 AND uc.UserId=$2 AND ca.AccessID<=$3
`

func (c *SessionClient) IsAuthorized(ctx context.Context, userID, companyID string, authorization AccessLevel) error {

	accessLevelID, err := authorization.ToAccessID()
	if err != nil {
		return errors.Wrap(err, "not able to get accesslevelid")
	}

	var userIDScan string
	err = c.DB.QueryRowContext(ctx, isAuthorizedQuery, companyID, userID, accessLevelID).Scan(&userIDScan)
	if err != nil {
		return errors.Wrap(err, "could not check if user is authorized")
	}

	if userIDScan != userID {
		return errors.Wrap(errors.New("invalid user id"), "suspicious")
	}

	return nil
}

func (c *SessionClient) IsAuthorizedFromEchoContext(ctx echo.Context, authorization AccessLevel) (authorized bool, err error) {
	companyID := ctx.Param("company")
	if companyID == "" {
		return false, errors.New("company id not provided")
	}

	userID := ctx.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		return false, errors.New("userid is not in scope")
	}

	err = c.IsAuthorized(ctx.Request().Context(), userID, companyID, authorization)
	if err != nil {
		return
	}

	return true, nil
}
