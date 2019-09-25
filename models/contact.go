package models

import (
	"github.com/jinzhu/gorm"
	"github.com/kristofhb/CreatixBackend/utils"
)

type Contact struct {
	gorm.Model
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	UserId int    `json;"user_id"`
}

func (contact *Contact) Validate() (map[string]interface{}, bool) {
	if contact.Name == "" {
		return utils.Message(false, "Contact name should be in the payload"), false
	}

	if contact.Phone == "" {
		return utils.Message(false, "Phone number should be in the payload"), false
	}

	if contact.UserId <= 0 {
		return utils.Message(false, "User is not recognized"), false
	}

	return utils.Message(true, "Success"), true
}
