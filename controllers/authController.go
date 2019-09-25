package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/kristofhb/CreatixBackend/models"
	"github.com/kristofhb/CreatixBackend/utils"
)

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	// Decode request body into Account struct
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		utils.Respond(w, utils.Message(false, "Invalid request"))
		return
	}

	// Create account
	resp := account.Create()
	utils.Respond(w, resp)

}

var Authenticate = func(w http.ResponseWriter, r *http.Request) {
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)

	if err != nil {
		utils.Respond(w, utils.Message(false, "Invalid request"))
		return
	}

	resp := models.Login(account.Email, account.Password)
	utils.Respond(w, resp)
}
