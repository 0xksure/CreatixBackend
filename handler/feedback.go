package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/logging"
)

type RestAPI struct {
	DB             *sql.DB
	Logging        *logging.StandardLogger
	Cfg            config.Config
	Feedback       models.Feedback
	Middleware     *middleware.Middleware
	CompanyClient  *models.CompanyClient
	SessionClient  *models.SessionClient
	FeedbackClient *models.FeedbackClient
}

var (
	PostFeedbackPath              = "/user/:company/feedback"
	GetFeedbackForUserCompanyPath = "/user/:company/feedback"
	DeleteFeedbackForUser         = "/feedback/:fid"
	PutFeedbackForUser            = "/feedback/:fid"
	PostClapFeedbackForUser       = "/user/feedback/:fid/clap"
	PostCommentFeedbackForUser    = "/user/feedback/:fid/comment"
)

func (api RestAPI) Handler(e *echo.Group) {
	e.Use(api.Middleware.JwtVerify)
	e.POST(PostFeedbackPath, api.PostFeedback)
	e.GET(GetFeedbackForUserCompanyPath, api.GetUserFeedback)

	e.DELETE(DeleteFeedbackForUser, api.DeleteFeedback)
	e.PUT(PutFeedbackForUser, api.UpdateFeedback)
	e.POST(PostClapFeedbackForUser, api.ClapFeedback)
	e.POST(PostCommentFeedbackForUser, api.CommentFeedback)

	api.CompanyHandler(e)

	e.GET("/ws/:company/feedback", api.FeedbackWebSocket)
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
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Write)
	if err != nil || !isAuthorized {

		api.Logging.Unsuccessful("creatix.feedback.PostFeedback: no permission", utils.NoPermission)
		return c.String(http.StatusUnauthorized, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.PostFeedback: no permission", utils.NoPermission)
		return c.String(http.StatusUnauthorized, "")
	}

	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	feedback := new(models.FeedbackRequest)
	if err = c.Bind(feedback); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no feedback provided", nil)
		return c.String(http.StatusBadRequest, "")
	}

	if err = api.FeedbackClient.CreateFeedback(c.Request().Context(), userID, companyID, *feedback); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to save feedback", err)
		return c.String(http.StatusInternalServerError, "")
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "posted feedback"})
}

// DeleteFeedback deletes feedback given an id
func (api RestAPI) DeleteFeedback(c echo.Context) error {
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Write)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.deletefeedback: user does not have permission", err)
		return errors.New("user does not have permission")
	}

	feedbackID := c.Param("fid")
	err = api.FeedbackClient.DeleteFeedback(c.Request().Context(), feedbackID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.deletefeedback: not able to delete feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to delete feedback"})
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})
}

// UpdateFeedback updates feedback based on the id in the url
func (api RestAPI) UpdateFeedback(c echo.Context) (err error) {
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Read)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no permission", err)
		return utils.NoPermission
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no permission", err)
		return utils.NoPermission
	}

	feedbackID := c.Param("fid")
	if feedbackID == "" {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to update feedback", err)
		return errors.New("no feedback id provided")
	}

	isOwner, err := api.FeedbackClient.IsUserOwnerOfFeedback(c.Request().Context(), feedbackID, userID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no permission", err)
		return err
	}

	if !isOwner {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no permission", err)
		return utils.NoPermission
	}

	feedback := new(models.FeedbackRequest)
	if err = c.Bind(feedback); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to update feedback", err)
		return err
	}

	err = api.FeedbackClient.UpdateFeedback(c.Request().Context(), feedbackID, *feedback)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: not able to update feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to update feedback"})
	}

	api.Logging.Success("creatix.feedback.updatefeedback: update feedback success")
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "ok"})

}

// ClapFeedback gives claps to a feedback given id
func (api RestAPI) ClapFeedback(c echo.Context) error {
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Read)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.ClapFeedback: no permission", err)
		return utils.NoPermission
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.ClapFeedback: no permission", nil)
		return utils.NoPermission
	}

	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("creatix.feedback.ClapFeedback: no company", nil)
		return c.String(http.StatusBadRequest, "")
	}

	feedbackID := c.Param("fid")
	if feedbackID == "" {
		api.Logging.Unsuccessful("creatix.feedback.updatefeedback: no feedback id provided", nil)
		return errors.New("no feedback id provided")
	}
	err = api.FeedbackClient.ClapFeedback(c.Request().Context(), userID, feedbackID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.clapfeedback: not able to clap feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to clap feedback"})
	}

	feedbacks, err := api.FeedbackClient.GetCompanyFeedbackswData(c.Request().Context(), companyID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback claps", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}

// GetUserFeedback gets all the feedback for the given user
func (api RestAPI) GetUserFeedback(c echo.Context) error {
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Read)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", utils.NoPermission)
		return c.String(http.StatusUnauthorized, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", utils.NoPermission)
		return c.String(http.StatusUnauthorized, "")
	}

	feedbacks, err := api.FeedbackClient.GetUserFeedbackwData(c.Request().Context(), userID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback", err)
		return c.String(http.StatusInternalServerError, "")
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
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Read)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", err)
		return utils.NoPermission
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", nil)
		return utils.NoPermission
	}

	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// SEND
		feedbacks, err := api.FeedbackClient.GetCompanyFeedbackswData(c.Request().Context(), companyID)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to get feedback", err)
			break
		}
		if err = ws.WriteJSON(feedbacks); err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to write feedback", err)
			break
		}

		// Receive
		var wsRequest models.WebSocketRequest
		err = ws.ReadJSON(&wsRequest)
		if err != nil {
			api.Logging.Unsuccessful("creatix.feedback.FeedbackWebsocket: not able to parse feedback", err)
			break
		}
		api.Logging.Success(fmt.Sprint("successfully parsed ", wsRequest.FeedbackID, wsRequest.Comment))
		switch wsRequest.Action {
		case 1:
			err = api.FeedbackClient.CreateFeedback(c.Request().Context(), userID, companyID, wsRequest.Feedback)
			if err != nil {
				api.Logging.Unsuccessful("creatix.feedback.CreateFeedback", err)
			}
		case 2:
			err = api.FeedbackClient.ClapFeedback(c.Request().Context(), userID, wsRequest.FeedbackID)
			if err != nil {
				api.Logging.Unsuccessful("creatix.feedback.ClapFeedback", err)
			}
		case 3:
			err = api.FeedbackClient.CommentFeedback(c.Request().Context(), wsRequest.Comment.Comment, userID, wsRequest.FeedbackID)
			if err != nil {
				api.Logging.Unsuccessful("creatix.feedback.CommentFeedback", err)
			}
		case 4:
			err = api.FeedbackClient.UpdateComment(c.Request().Context(), wsRequest.Comment.ID, wsRequest.Comment.Comment)
			if err != nil {
				api.Logging.Unsuccessful("creatix.feedback.UpdateComment", err)
			}
		default:
			api.Logging.Unsuccessful(fmt.Sprintf("creatix.feedback.FeedbackWebsocket: option %d is not a valid ws option", wsRequest.Action), err)
			break
		}

	}
	return nil
}

// CommentFeedback comments a given feedback
func (api RestAPI) CommentFeedback(c echo.Context) (err error) {
	isAuthorized, err := api.SessionClient.IsAuthorizedFromEchoContext(c, models.Read)
	if err != nil || !isAuthorized {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", err)
		return utils.NoPermission
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.getuserfeedback: no permission", nil)
		return utils.NoPermission
	}

	comment := new(models.CommentRequest)
	if err = c.Bind(comment); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to bind comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to bind comment"})
	}

	feedbackID := c.Param("fid")
	if err = api.FeedbackClient.CommentFeedback(c.Request().Context(), comment.Comment, userID, feedbackID); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.commentfeedback: not able to update comment", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to write comment"})
	}

	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("creatix.feedback.ClapFeedback: no company", nil)
		return c.String(http.StatusBadRequest, "")
	}

	feedbacks, err := api.FeedbackClient.GetCompanyFeedbackswData(c.Request().Context(), companyID)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.getUserFeedback: not able to get feedback claps", err)
		return err
	}
	return c.JSON(http.StatusOK, feedbacks)
}
