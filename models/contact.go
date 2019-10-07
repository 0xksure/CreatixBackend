package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kristofhb/CreatixBackend/utils"
)

func SignUpNewsletter(nl *Newsletter) (map[string]interface{}, error) {
	fmt.Println(nl)
	if resp, ok := ValidateNewsletter(nl); !ok {
		return resp, errors.New("Could not validate")
	}
	GetDB().Create(nl)

	return utils.Message(false, "Signup complete"), nil
}

//ValidateNewsletter validates the input from the request
func ValidateNewsletter(nl *Newsletter) (map[string]interface{}, bool) {
	selectedProducts := []int{1, 2, 3}
	// Check if selectedproduct is specified
	if !utils.SliceContains(selectedProducts, nl.SelectedProduct) {
		return utils.Message(false, "Selected Product is not valid"), false
	}

	// Check if email is valid
	fmt.Println("Email: ", nl.Email)
	if !strings.Contains(nl.Email, "@") {
		return utils.Message(false, "Email address is required"), false
	}

	return utils.Message(false, "Requirement passed"), true

}
