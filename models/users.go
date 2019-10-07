package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

//JWT claims structure
type Token struct {
	UserID uint
	jwt.StandardClaims
}

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// User Accounts
type Account struct {
	gorm.Model
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Token    string    `json:"token"; sql:"-"`
	Expires  time.Time `json:"expires"`
}

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Contact struct {
	gorm.Model
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	UserID int    `json:"userId"`
}

type Newsletter struct {
	gorm.Model
	SelectedProduct int    `json:"selectedProduct"`
	Email           string `json:"email"`
}
