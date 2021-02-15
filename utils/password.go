package utils

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type PasswordRequirements struct {
	ContainsUppercase bool
	ContainsLowercase bool
	ContainsDigit     bool
	SpecialLetter     bool
}

var specialLetters = []string{"@", "#"}

func (p PasswordRequirements) AllGood() (bool, error) {

	if !p.ContainsDigit {
		return false, errors.New("password does not contain digit")
	}

	if !p.ContainsUppercase {
		return false, errors.New("password does not contain an uppercase letter")
	}

	if !p.ContainsLowercase {
		return false, errors.New("password does not contain an lowercase letter")
	}

	if !p.SpecialLetter {

		return false, errors.New(fmt.Sprintf("password does not contain %s", strings.Join(specialLetters, ", ")))
	}

	return true, nil

}

func IsValidPassword(password string) (bool, error) {
	if len(password) <= 12 {
		return false, errors.New("password needs to be longer than 12 characters long")
	}

	var passwordRequirements PasswordRequirements
	for _, char := range password {
		if unicode.IsUpper(char) {
			passwordRequirements.ContainsUppercase = true
		}

		if unicode.IsLower(char) {
			passwordRequirements.ContainsLowercase = true
		}

		if unicode.IsNumber(char) {
			passwordRequirements.ContainsDigit = true
		}

		for _, specialChar := range specialLetters {
			if specialChar == string(char) {
				passwordRequirements.SpecialLetter = true
			}
		}
	}

	return passwordRequirements.AllGood()

}
