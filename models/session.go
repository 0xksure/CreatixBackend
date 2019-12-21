package models

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	Name     string
	Email    string `gorm:"type:varchar(100);unique_index"`
	Gender   string `json:"gender"`
	Password string `json:"password"`
}

type UserSession struct {
	UserID uint
	Name   string
	Email  string
	*jwt.StandardClaims
}

type Response struct {
	Status  bool
	Message string
	Token   string
	User
}

func CreateUser(db *gorm.DB, user User) (createdUser *gorm.DB, err error) {
	createdUser = db.Create(user)
	if err := createdUser.Error; err != nil {
		return createdUser, err
	}
	return createdUser, nil
}

func Login(user *User) (int, []byte) {
	authBackend := authentication.InitJWT
}

func LoginUser(db *gorm.DB, user User) (resp Response, err error) {
	var authUser User
	resp.User = user
	if err = db.Where("Email = ?", user.Email).First(authUser).Error; err != nil {
		return
	}

	expiresAt := time.Now().Add(time.Minute * 100000).Unix()
	errf := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authUser.Password))
	if errf == bcrypt.ErrMismatchedHashAndPassword {
		return resp, errors.New("Passwords do not match")
	}
	us := UserSession{
		UserID: authUser.ID,
		Name:   authUser.Name,
		Email:  authUser.Email,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), us)

	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		return
	}
	
	resp.Status = false
	resp.Message = "logged in"
	resp.Token = tokenString
	resp.User = authUser
	return resp, nil

}
