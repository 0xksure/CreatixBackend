package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/app"
	"github.com/kristofhb/CreatixBackend/controllers"
)

func Restapi() {

	router := mux.NewRouter().StrictSlash(true)
	router.Use(app.JwtAuthentication)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/api/event", createEvent)
	router.HandleFunc("/api/event/{id}", getOneEvent)
	router.HandleFunc("/api/events", getAllEvents)
	router.HandleFunc("/api/user/new", controllers.CreateAccount).Methods("POST")
	router.HandleFunc("/api/user/login", controllers.Authenticate).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome Home")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var newEvent event
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter title and description")
	}

	json.Unmarshal(reqBody, &newEvent)
	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}
