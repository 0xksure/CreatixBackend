package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/models"
	"github.com/kristofhb/CreatixBackend/utils"
)

// NewsletterAPI is responsible for catching everything that is
// related to signing up for more contact
func (a *RestAPI) NewsletterAPI(r *mux.Router) {
	r.HandleFunc("/api/v1/newsletter/signup", a.signUpNewsletter)
}

// newsletter responds on sent in request to join newsletter
func (a *RestAPI) signUpNewsletter(w http.ResponseWriter, r *http.Request) {
	newsletter := &models.Newsletter{}
	err := json.NewDecoder(r.Body).Decode(newsletter)
	if err != nil {
		utils.Respond(w, utils.Message(false, "Invalid request"))
		w.WriteHeader(http.StatusUnauthorized)
	}
	resp, err := models.SignUpNewsletter(newsletter)

	utils.Respond(w, resp)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
	}

	return
}
