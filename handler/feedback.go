package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/logging"
)

type RestAPI struct {
	DB            *sql.DB
	Logging       *logging.StandardLogger
	Cfg           config.Config
	Feedback      models.Feedback
	Middleware    *middleware.Middleware
	CompanyClient models.CompanyClient
}

func (api RestAPI) Handler(e *echo.Group) {
	e.Use(api.Middleware.JwtVerify)
	e.POST("/user/feedback", api.PostFeedback)
	e.GET("/user/feedback", api.GetUserFeedback)
	e.DELETE("/feedback/:fid", api.DeleteFeedback)
	e.PUT("/feedback", api.UpdateFeedback)
	e.POST("/user/feedback/:fid/clap", api.ClapFeedback)
	e.POST("/user/feedback/:fid/comment", api.CommentFeedback)
	e.GET("/company/search/{query}", api.SearchCompany)
	e.POST("/company/create", api.CreateCompany)
	e.GET("/ws/feedback", api.FeedbackWebSocket)
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

func (api RestAPI) postFeedback(feedback *models.Feedback, ctx context.Context) (err error) {
	feedback.UserID = api.Middleware.Uid
	if err = feedback.CreateFeedback(ctx, api.DB); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to save feedback", err)
		return
	}
	return nil
}

// PostFeedback posts feedback from user
func (api RestAPI) PostFeedback(c echo.Context) (err error) {
	feedback := new(models.Feedback)
	if err = c.Bind(feedback); err != nil {
		return
	}
	err = api.postFeedback(feedback, c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to post feedback"})

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
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}

var upgrader = websocket.Upgrader{ReadBufferSize: 4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func (api RestAPI) FeedbackWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// SEND
		feedbacks, err := api.getUserFeedback(c.Request().Context(), api.Middleware.Uid)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to get feedback", err)
			break
		}
		if err = ws.WriteJSON(feedbacks); err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to write feedback", err)
			break
		}

		// Receive
		wsRequest := models.WebSocketRequest{}
		err = ws.ReadJSON(&wsRequest)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to parse feedback", err)

		}
		switch wsRequest.Action {
		case 1:
			api.postFeedback(&wsRequest.Feedback, c.Request().Context())
		case 2:
			api.Feedback.ClapFeedback(c.Request().Context(), api.DB, api.Middleware.Uid, wsRequest.FeecbackID)
		case 3:
			api.commentFeedback(wsRequest.Comment, c.Request().Context())
		default:
			api.Logging.Unsuccessful(fmt.Sprintf("creatix.feedback.FeedbackWebsocket: option %d is not a valid ws option", wsRequest.Action), err)
			break
		}

	}
	return nil
}

func (api RestAPI) commentFeedback(comment models.Comment, ctx context.Context) (err error) {
	comment.UserID = api.Middleware.Uid

	if comment.ID != "" {
		err = comment.UpdateComment(ctx, api.DB)
		if err != nil {
			return err
		}
	} else {
		err = comment.CommentFeedback(ctx, api.DB, api.Middleware.Uid)
		if err != nil {
			return err
		}
	}
	return nil
}

// CommentFeedback comments a given feedback
func (api RestAPI) CommentFeedback(c echo.Context) (err error) {
	comment := new(models.Comment)
	if err = c.Bind(comment); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to bind comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to bind comment"})
	}
	comment.FeedbackID = c.Param("fid")
	if err = api.commentFeedback(*comment, c.Request().Context()); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to update comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to write comment"})
	}

	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})
}
