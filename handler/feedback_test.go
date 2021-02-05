package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/test"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFeedback(ID, userID, firstname, lastname string) []models.Feedback {
	return []models.Feedback{{
		ID:          ID,
		UserID:      userID,
		Person:      models.Person{ID: "", Firstname: firstname, Lastname: lastname},
		Title:       "This is my title",
		Description: "A title says more than a thousand words",
		Comments:    []models.Comment{},
		Claps:       []models.Clap{},
		UpdatedAt:   nil,
	}}
}

func newFeedbackRequest() models.FeedbackRequest {
	return models.FeedbackRequest{
		Title:       "This is my title",
		Description: "A title says more than a thousand words",
	}
}

func NewMockFeedbackBytes() (feedback []byte, err error) {
	return json.Marshal(newFeedbackRequest())
}

func SetupRequest(c echo.Context, userID, companyID string) {
	c.SetPath("/company/:company/adduser")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
}

func NewRequester(t *testing.T, e *echo.Echo, restAPI RestAPI) func(path, userID string, parameterNameValue [][]string, load []byte, f func(echo.Context) error, expectedCode int, expectedBody interface{}) {
	return func(path, userID string, parameterNameValue [][]string, load []byte, f func(echo.Context) error, expectedCode int, expectedBody interface{}) {
		c, rec := newContext(e, load, path)
		c.SetPath(path)

		c.SetParamNames(parameterNameValue[0]...)
		c.SetParamValues(parameterNameValue[1]...)

		c.Set(utils.UserIDContext.String(), userID)

		err := f(c)
		if assert.NoError(t, err) {
			assert.Equal(t, expectedCode, rec.Code)

			if expectedBody != nil {
				expectedByte, err := json.Marshal(expectedBody)
				require.NoError(t, err)
				assert.Equal(t, string(expectedByte), getStringJSON(rec))
			}

		}
	}
}

func TestFeedback(t *testing.T) {
	// Setup
	e := echo.New()

	db, err := test.NewTestDB()
	require.NoError(t, err)
	err = test.TestMigrations(db)
	require.NoError(t, err)
	defer test.EmptyTestDB(t, db)

	// Create user
	logger := logging.NewLogger()
	restAPI := NewRestAPI(db, logger)
	requester := NewRequester(t, e, restAPI)

	// Add New Admin user
	parameterNameValue := [][]string{{"company"}, {"1"}}
	newAdminUser, err := NewMockUserBytes("john@doe.no", models.Admin)
	require.NoError(t, err)
	requester(AddNewUserByEmailToCompanyPath, "1", parameterNameValue, newAdminUser, restAPI.AddUserByEmailToCompany, http.StatusOK, nil)

	// Postfeedback as admin
	feedback, err := NewMockFeedbackBytes()
	require.NoError(t, err)
	requester(PostFeedbackPath, "1", parameterNameValue, feedback, restAPI.PostFeedback, http.StatusOK, nil)

	// Postfeedback as write access
	feedback, err = NewMockFeedbackBytes()
	require.NoError(t, err)
	requester(PostFeedbackPath, "2", parameterNameValue, feedback, restAPI.PostFeedback, http.StatusOK, nil)

	// Postfeedback with read access
	feedback, err = NewMockFeedbackBytes()
	require.NoError(t, err)
	requester(PostFeedbackPath, "3", parameterNameValue, feedback, restAPI.PostFeedback, http.StatusUnauthorized, nil)

	// GET feedback as admin
	requester(GetFeedbackForUserCompanyPath, "1", parameterNameValue, nil, restAPI.GetUserFeedback, http.StatusOK, newFeedback("1", "1", "Kristoffer", "Berg"))

	// GET feedback as Reader
	requester(GetFeedbackForUserCompanyPath, "3", parameterNameValue, nil, restAPI.GetUserFeedback, http.StatusUnauthorized, nil)

	// Delete Feedback
	parameterNameValue[0] = append(parameterNameValue[0], "fid")
	parameterNameValue[1] = append(parameterNameValue[1], "1")
	requester(DeleteFeedbackForUser, "1", parameterNameValue, nil, restAPI.DeleteFeedback, http.StatusOK, nil)

	// GET feedback - will return empty
	requester(GetFeedbackForUserCompanyPath, "1", parameterNameValue, nil, restAPI.GetUserFeedback, http.StatusOK, nil)
}

func TestClapFeedback(t *testing.T) {
	e := echo.New()

	db, err := test.NewTestDB()
	require.NoError(t, err)
	err = test.TestMigrations(db)
	require.NoError(t, err)
	defer test.EmptyTestDB(t, db)

	// Create user
	logger := logging.NewLogger()
	restAPI := NewRestAPI(db, logger)
	requester := NewRequester(t, e, restAPI)

	// New reader user
	parameterNameValue := [][]string{{"company"}, {"1"}}
	newReaderUser, err := NewMockUserBytes("john@doe.no", models.Read)
	require.NoError(t, err)
	requester(AddNewUserByEmailToCompanyPath, "1", parameterNameValue, newReaderUser, restAPI.AddUserByEmailToCompany, http.StatusOK, nil)

	// Postfeedback as admin
	feedback, err := NewMockFeedbackBytes()
	require.NoError(t, err)
	requester(PostFeedbackPath, "1", parameterNameValue, feedback, restAPI.PostFeedback, http.StatusOK, nil)

	// Clap feedback
	clappedFeedback := models.Feedbacks{
		models.Feedback{ID: "1", UserID: "1",
			Person: models.Person{
				ID:        "",
				Firstname: "Kristoffer",
				Lastname:  "Berg",
			},
			Title:       "This is my title",
			Description: "A title says more than a thousand words",
			Comments:    []models.Comment{},
			Claps:       []models.Clap{{ID: "1", UserID: "2", FeedbackID: "1"}},
			UpdatedAt:   nil}}
	parameterNameValue[0] = append(parameterNameValue[0], "fid")
	parameterNameValue[1] = append(parameterNameValue[1], "1")
	requester(PostClapFeedbackForUser, "2", parameterNameValue, nil, restAPI.ClapFeedback, http.StatusOK, clappedFeedback)

	// unclap feedback
	clappedFeedback[0].Claps = []models.Clap{}
	requester(PostClapFeedbackForUser, "2", parameterNameValue, nil, restAPI.ClapFeedback, http.StatusOK, clappedFeedback)

	// Delete feedback
	requester(DeleteFeedbackForUser, "1", parameterNameValue, nil, restAPI.DeleteFeedback, http.StatusOK, nil)

	// Try to clap nonexisting feedback
	requester(PostClapFeedbackForUser, "2", parameterNameValue, nil, restAPI.ClapFeedback, http.StatusOK, nil)
}

func TestCommentFeedback(t *testing.T) {
	e := echo.New()

	db, err := test.NewTestDB()
	require.NoError(t, err)
	err = test.TestMigrations(db)
	require.NoError(t, err)
	defer test.EmptyTestDB(t, db)

	// Create user
	logger := logging.NewLogger()
	restAPI := NewRestAPI(db, logger)
	requester := NewRequester(t, e, restAPI)

	// New reader user
	parameterNameValue := [][]string{{"company"}, {"1"}}
	newReaderUser, err := NewMockUserBytes("john@doe.no", models.Read)
	require.NoError(t, err)
	requester(AddNewUserByEmailToCompanyPath, "1", parameterNameValue, newReaderUser, restAPI.AddUserByEmailToCompany, http.StatusOK, nil)

	// Postfeedback as admin
	feedback, err := NewMockFeedbackBytes()
	require.NoError(t, err)
	requester(PostFeedbackPath, "1", parameterNameValue, feedback, restAPI.PostFeedback, http.StatusOK, nil)

	// Comment feedback
	newComment := models.CommentRequest{Comment: "wow cool"}
	newCommentByte, err := json.Marshal(newComment)
	commentedFeedback := models.Feedbacks{
		models.Feedback{
			ID:     "1",
			UserID: "1",
			Person: models.Person{
				ID:        "",
				Firstname: "Kristoffer",
				Lastname:  "Berg",
			},
			Title:       "This is my title",
			Description: "A title says more than a thousand words",
			Comments:    []models.Comment{{ID: "1", FeedbackID: "1", Person: models.Person{ID: "2", Firstname: "John", Lastname: "Doe"}, Comment: newComment.Comment}},
			Claps:       []models.Clap{},
			UpdatedAt:   nil,
		}}
	require.NoError(t, err)
	parameterNameValue[0] = append(parameterNameValue[0], "fid")
	parameterNameValue[1] = append(parameterNameValue[1], "1")
	requester(PostCommentFeedbackForUser, "2", parameterNameValue, newCommentByte, restAPI.CommentFeedback, http.StatusOK, commentedFeedback)
}
