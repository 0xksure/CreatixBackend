package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	jwtmiddleware "github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/logging"
)

type RestAPI struct {
	DB          *sql.DB
	Logging     *logging.StandardLogger
	Cfg         config.Config
	Feedback    models.Feedback
	UserSession *models.UserSession
	Middleware  *jwtmiddleware.Middleware
}

func (api RestAPI) Handler(e *echo.Group) {
	e.Use(api.Middleware.JwtVerify)
	e.POST("/feedback", api.PostFeedback)
	e.GET("/user/feedback", api.GetUserFeedback)
	e.DELETE("/feedback/:fid", api.DeleteFeedback)
	e.PUT("/feedback", api.UpdateFeedback)
	e.POST("/user/feedback/:fid/clap", api.ClapFeedback)
	e.POST("/user/feedback/comment", api.CommentFeedback)
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
func (api RestAPI) PostFeedback(c echo.Context) (err error) {
	feedback := new(models.Feedback)
	if err = c.Bind(feedback); err != nil {
		return
	}
	feedback.UserID = api.Middleware.Uid
	if err = feedback.CreateFeedback(c.Request().Context(), api.DB); err != nil {
		fmt.Println("error in post feedback")
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to save feedback", err)
		return
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "posted feedback"})
}

// DeleteFeedback deletes feedback given an id
func (api RestAPI) DeleteFeedback(c echo.Context) error {
	feedbackID := c.Param("fid")
	err := api.Feedback.DeleteFeedback(c.Request().Context(), api.DB, feedbackID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.deletefeedback: not able to delete feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to delete feedback"})
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})
}

// UpdateFeedback updates feedback based on the id in the url
func (api RestAPI) UpdateFeedback(c echo.Context) (err error) {
	feedback := new(models.Feedback)
	if err = c.Bind(feedback); err != nil {
		return err
	}
	err = feedback.UpdateFeedback(c.Request().Context(), api.DB)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to update feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to update feedback"})
	}

	api.Logging.Success("creatix.feedback.updatefeedback: update feedback success")
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})

}

// ClapFeedback gives claps to a feedback given id
func (api RestAPI) ClapFeedback(c echo.Context) error {
	feedbackID := c.Param("fid")
	err := api.Feedback.ClapFeedback(c.Request().Context(), api.DB, api.Middleware.Uid, feedbackID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.clapfeedback: not able to clap feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to delete feedback"})
	}
	feedbacks, err := api.Feedback.GetUserFeedback(c.Request().Context(), api.DB, api.UserSession.ID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.clapFeedback: not able to get feedbacks", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to delete feedback"})
	}
	api.Logging.Success("creatix.feedback.clapfeedback: successfully clapped feedback")

	return c.JSON(http.StatusOK, feedbacks)
}

// GetUserFeedback gets all the feedback for the given user
func (api RestAPI) GetUserFeedback(c echo.Context) error {
	feedbacks, err := api.Feedback.GetUserFeedback(c.Request().Context(), api.DB, api.Middleware.Uid)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedbacks", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}

// Comment feedback
func (api RestAPI) CommentFeedback(c echo.Context) (err error) {
	var comment models.Comment
	if err = c.Bind(comment); err != nil {
		return err
	}
	err = comment.CommentFeedback(c.Request().Context(), api.DB, api.Middleware.Uid)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to write comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to delete feedback"})
	}
	return c.JSON(http.StatusOK, nil)
}
