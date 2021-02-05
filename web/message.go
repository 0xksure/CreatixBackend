package web

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kristohberg/CreatixBackend/utils"
)

type HttpResponse struct {
	Message string
	User    utils.SessionUser
}

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Fatal("Ohh noo")
	}
}

func SliceContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
