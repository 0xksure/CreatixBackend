package models

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/kristofhb/CreatixBackend/utils"
	"golang.org/x/crypto/bcrypt"
)

// Validate incoming user details
func ValidateAccount(account *Account) (map[string]interface{}, bool) {
	if !strings.Contains(account.Email, "@") {
		return utils.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return utils.Message(false, "Password is required"), false
	}

	temp := &Account{}

	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.Message(false, "Connection error. Please retry"), false
	}

	if temp.Email != "" {
		return utils.Message(false, "Email address is allready in use"), false
	}

	return utils.Message(false, "Requirement passed"), true
}

func Create(account *Account, w http.ResponseWriter) map[string]interface{} {

	if resp, ok := ValidateAccount(account); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	GetDB().Create(account)

	if account.ID <= 0 {
		return utils.Message(false, "Failed to create account, connection error")
	}

	// Create new JWT token for the newly registered account

	// Expiration time of token
	expirationTime := time.Now().Add(5 * time.Minute)
	account.Expires = expirationTime

	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	account.Password = "" //delete password

	response := utils.Message(true, "Account has been created")
	response["account"] = account

	// Set the client cookie for token
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	return response
}

func Login(email string, password string) map[string]interface{} {

	account := &Account{}
	err := GetDB().Table("accounts").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Message(false, "Email can not be found")
		}
		return utils.Message(false, "Connection error. Please try again")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return utils.Message(false, "Invalid login credentials. Please try again")
	}

	// Worked logging in
	account.Password = ""

	// Create JWT token
	// Expiration time of token
	expirationTime := time.Now().Add(5 * time.Minute)
	account.Expires = expirationTime
	// Create the JWT claims,
	tk := &Token{UserID: account.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		}}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("token_password")))
	if err != nil {
		return utils.Message(false, "Invalid Login credentials. Please try again.")
	}
	account.Token = tokenString

	resp := utils.Message(true, "Logged in")
	resp["account"] = account

	// Set the client cookie for token
	/*http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})*/
	return resp
}

func GetToken(tokenHeader string) ([]string, map[string]interface{}, error) {
	fmt.Println("GET token")
	// token header is missing
	if tokenHeader == "" {
		return nil, utils.Message(false, "Invalid/Malformed auth token"), errors.New("Invalid Token")
	}
	fmt.Println("Token header exists")

	splitted := strings.Split(tokenHeader, " ")
	// Check if token comes in format `Bearer {token}`
	if len(splitted) != 2 {
		return nil, utils.Message(false, "Invalid/Malformed auth token"), errors.New("Missing Token")
	}

	return splitted, nil, nil
}

func GetUser(u uint) *Account {

	acc := &Account{}
	GetDB().Table("accounts").Where("id = ?", u).First(acc)
	if acc.Email == "" {
		return nil
	}

	acc.Password = ""
	return acc
}
