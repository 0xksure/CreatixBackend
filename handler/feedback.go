package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kristofhb/CreatixBackend/models"
	"github.com/kristofhb/CreatixBackend/utils"
	"github.com/labstack/echo"

	"github.com/gorilla/mux"
	"github.com/kristofhb/CreatixBackend/config"
	"github.com/kristofhb/CreatixBackend/logging"
	"github.com/kristofhb/CreatixBackend/middleware"
)

type RestAPI struct {
	DB       *sql.DB
	Logging  *logging.StandardLogger
	Cfg      *config.Config
	Feedback models.Feedback
}

func (api RestAPI) Handler(e *echo.Group) {
	e.Use(middleware.JwtVerify)
	e.POST("/feedback", api.PostFeedback)
	e.GET("/user/feedback", api.GetUserFeedback)
	e.DELETE("/feedback/{fid}", api.DeleteFeedback)
	e.PUT("/feedback/{fid}", api.UpdateFeedback)
	e.POST("/user/feedback/{fid}/clap", api.ClapFeedback)
	e.POST("/user/feedback/{fid}/comment", api.CommentFeedback)
}

func validateFeedback(feedback models.Feedback) error {
	if feedback.Title == "" {
		return errors.New("creatix.feedback.validate: feedback title is to short")
	}
	if feedback.Description == "" {
		return errors.New("creatix.feedback.validate: feedback description is too short")
	}

	return nil
}

// PostFeedback posts feedback from user
func (api RestAPI) PostFeedback(c echo.Context) error {
	var feedback models.Feedback
	err := c.Bind(feedback)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to parse feedback", err)
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "not able to parse feedback"})
	}

	err = feedback.CreateFeedback(c.Request().Context(), api.DB)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to save feedback", err)
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "not able to save feedback"})
	}

	return c.JSON(http.StatusOK, utils.HttpResponse{Message: "ok"})
}

// DeleteFeedback deletes feedback given an id
func (api RestAPI) DeleteFeedback(c echo.Context) error {
	feedbackID := c.Param("fid")
	err := api.Feedback.DeleteFeedback(c.Request().Context(), api.DB, feedbackID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.deletefeedback: not able to delete feedback", err)
		return c.JSON(http.StatusBadRequest, utils.HttpResponse{Message: "not able to delete feedback"})
	}
	return c.JSON(http.StatusOK, utils.HttpResponse{Message: "ok"})
}

// UpdateFeedback updates feedback based on the id in the url
func (api RestAPI) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("User Email= ", api.User.Email)
	feedbackID := mux.Vars(r)["fid"]
	feedback := api.Feedback
	err := json.NewDecoder(r.Body).Decode(&feedback)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to decode feedback", err)
		return
	}
	updatedFeedback, err := feedback.UpdateFeedback(api.DB, api.User, feedbackID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to update feedback", err)
		return
	}

	err = json.NewEncoder(w).Encode(updatedFeedback)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to write out updated feedback", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	api.Logging.Success("creatix.feedback.updatefeedback: update feedback success")
}

// ClapFeedback gives claps to a feedback given id
func (api RestAPI) ClapFeedback(w http.ResponseWriter, r *http.Request) {
	feedbackID := mux.Vars(r)["fid"]
	userEmail := api.User.Email
	if userEmail == "" {
		w.WriteHeader(http.StatusUnauthorized)
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to find credentials", errors.New("unauthorized"))
		return
	}
	_, err := api.Feedback.ClapFeedback(api.DB, userEmail, feedbackID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.clapfeedback: not able to clap feedback", err)
		return
	}
	feedbacks, err := api.Feedback.GetUserFeedback(api.DB, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.clapFeedback: not able to get feedbacks", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	api.Logging.Success("creatix.feedback.clapfeedback: successfully clapped feedback")
	err = json.NewEncoder(w).Encode(feedbacks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to encode feedbacks", err)
		return
	}
}

// GetUserFeedback gets all the feedback for the given user
func (api RestAPI) GetUserFeedback(w http.ResponseWriter, r *http.Request) {
	userEmail := api.User.Email
	if userEmail == "" {
		w.WriteHeader(http.StatusUnauthorized)
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to find credentials", errors.New("unauthorized"))
		return
	}
	feedbacks, err := api.Feedback.GetUserFeedback(api.DB, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedbacks", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(feedbacks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to encode feedbacks", err)
		return
	}
}

// Comment feedback
func (api RestAPI) CommentFeedback(w http.ResponseWriter, r *http.Request) {
	var comment models.Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to decode comment", err)
		return
	}
	feedbackID := mux.Vars(r)["fid"]
	userEmail := api.User.Email
	if userEmail == "" {
		w.WriteHeader(http.StatusUnauthorized)
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to find credentials", errors.New("unauthorized"))
		return
	}

	err = api.Feedback.CommentFeedback(api.DB, userEmail, feedbackID, comment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to write comment", err)
		return
	}
	w.WriteHeader(http.StatusOK)

}
