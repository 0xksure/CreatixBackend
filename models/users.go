package models

import (
	"context"
	"fmt"
)

type AccessLevel string

type AccessID int

const (
	AdminID AccessID = iota
	WriteID
	ReadID
)

const (
	Admin AccessLevel = "admin"
	Write             = "write"
	Read              = "read"
)

// ToAccessID converts accesslevel to accessid
func (a AccessLevel) ToAccessID() (AccessID int, err error) {
	switch a {
	case "admin":
		return 1, nil
	case "write":
		return 2, nil
	case "read":
		return 3, nil
	default:
		return 0, fmt.Errorf("%s is not a valid access level", a)
	}

}

type AddUser struct {
	Email  string      `json:"email"`
	Access AccessLevel `json:"accessLevel"`
}

type UserPermissionRequest struct {
	UserID int         `json:"userId"`
	Access AccessLevel `json:"accessLevel"`
}

type CompanyUserResponse struct {
	UserID   int         `json:"userId"`
	Username string      `json:"username"`
	Access   AccessLevel `json:"accessLevel"`
}
type UserClient interface {
	InviteUserByEmail(ctx context.Context) error
}
