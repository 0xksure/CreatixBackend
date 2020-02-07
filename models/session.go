package models

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
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

type UserSession struct {
	UserID    uint
	Firstname string
	Lastname  string
	Email     string
	*jwt.StandardClaims
}

type Response struct {
	Status    bool
	Message   string
	Token     string
	ExpiresAt time.Time
	User
}

func (user *User) CreateUser(db *gorm.DB) (createdUser *gorm.DB, err error) {
	createdUser = db.Create(&user)
	if err := createdUser.Error; err != nil {
		return createdUser, err
	}
	return createdUser, nil
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

func (user *User) LoginUser(db *gorm.DB) (resp Response, err error) {
	var authUser User
	resp.User = *user
	if err = db.Where("Email = ?", user.Email).First(&authUser).Error; err != nil {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, err
	}
	errf := bcrypt.CompareHashAndPassword([]byte(authUser.Password), []byte(user.Password))
	if errf == bcrypt.ErrMismatchedHashAndPassword {
		resp.Message = "Either the user does not exists or the password is incorrect"
		return resp, errors.New("Passwords do not match")
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
