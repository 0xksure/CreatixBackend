package handler

import (
	"context"
	"database/sql"
	"errors"
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
	CompanyAPI  models.CompanyAPI
}

func (api RestAPI) Handler(e *echo.Group) {
	e.Use(api.Middleware.JwtVerify)
	e.POST("/feedback", api.PostFeedback)
	e.GET("/user/feedback", api.GetUserFeedback)
	e.DELETE("/feedback/:fid", api.DeleteFeedback)
	e.PUT("/feedback", api.UpdateFeedback)
	e.POST("/user/feedback/:fid/clap", api.ClapFeedback)
	e.POST("/user/feedback/comment", api.CommentFeedback)
	e.GET("/company/search/{query}", api.SearchCompany)
	e.POST("/company/create", api.CreateCompany)
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
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to clap feedback"})
	}
	api.Logging.Success("creatix.feedback.clapfeedback: successfully clapped feedback")

	feedbacks, err := api.getUserFeedback(c.Request().Context(), api.Middleware.Uid)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback claps", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}

func (api RestAPI) getUserFeedback(ctx context.Context, userID string) (feedbacks models.Feedbacks, err error) {
	feedbacks, err = api.Feedback.GetUserFeedback(ctx, api.DB, userID)
	if err != nil {
		return
	}

	err = feedbacks.GetUserComments(ctx, api.DB)
	if err != nil {
		return
	}

	err = feedbacks.GetUserClaps(ctx, api.DB)
	if err != nil {
		return
	}
	return feedbacks, nil
}

// GetUserFeedback gets all the feedback for the given user
func (api RestAPI) GetUserFeedback(c echo.Context) error {
	feedbacks, err := api.getUserFeedback(c.Request().Context(), api.Middleware.Uid)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback claps", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}

// CommentFeedback comments a given feedback
func (api RestAPI) CommentFeedback(c echo.Context) (err error) {
	comment := new(models.Comment)
	if err = c.Bind(comment); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to bind comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to bind comment"})
	}

	if comment.ID != "" {
		err = comment.UpdateComment(c.Request().Context(), api.DB)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to update comment", err)
			return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to update comment"})
		}
	} else {
		err = comment.CommentFeedback(c.Request().Context(), api.DB, api.Middleware.Uid)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to write comment", err)
			return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to write comment"})
		}
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})
}
