package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/middleware"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/test"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func InitConfig() config.Config {
	return config.Config{Env: "test"}
}

func SignupAndLoginUser(t *testing.T, e *echo.Echo, sessionAPI SessionAPI, logger *logging.StandardLogger) (cookie string) {
	var c echo.Context
	var rec *httptest.ResponseRecorder

	signupLoad, err := NewSignupUserByte("john", "doe", "johndoe1", "john@doe.com", "lol")
	require.NoError(t, err)
	c, rec = newContext(e, signupLoad, "")
	// SIGNUP USER
	res := sessionAPI.Signup(c)
	require.NoError(t, res)

	// Login user
	loginRequestByte, err := json.Marshal(newLoginRequest(NewSignupUser("john", "doe", "johndoe1", "john@doe.com", "lol")))
	require.NoError(t, err)
	c, rec = newContext(e, loginRequestByte, "")
	err = sessionAPI.Login(c)
	require.NoError(t, err)

	// get cookie

	return rec.Header().Get("Set-Cookie")
}

func NewRestAPI(db *sql.DB, logger *logging.StandardLogger) RestAPI {
	cfg := InitConfig()
	return RestAPI{
		DB:             db,
		Logging:        logger,
		Cfg:            cfg,
		Feedback:       models.Feedback{},
		Middleware:     &middleware.Middleware{Cfg: cfg},
		CompanyClient:  models.NewCompanyClient(db),
		SessionClient:  models.NewSessionClient(db, []byte("test"), 30, logger),
		FeedbackClient: models.NewFeedbackClient(db),
	}
}

func NewSessionAPI(db *sql.DB, logger *logging.StandardLogger) SessionAPI {
	return SessionAPI{
		DB:            db,
		Logging:       logger,
		Cfg:           config.Config{Env: "test"},
		SessionClient: models.NewSessionClient(db, []byte("secret"), 20, logger),
	}
}

func TestCompany(t *testing.T) {
	// Setup
	var c echo.Context
	var rec *httptest.ResponseRecorder
	e := echo.New()

	db, err := test.NewTestDB()
	require.NoError(t, err)
	err = test.TestMigrations(db)
	require.NoError(t, err)
	defer test.EmptyTestDB(t, db)

	// Create user
	logger := logging.NewLogger()
	restAPI := NewRestAPI(db, logger)

	// Create new company
	companyJSON, err := json.Marshal(newCompany())
	require.NoError(t, err)
	c, rec = newContext(e, companyJSON, "/company/create")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.CreateCompany(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Add New user
	newUserLoad, err := json.Marshal(models.AddUser{Email: "john@doe.no", Access: models.Write})
	require.NoError(t, err)
	c, rec = newContext(e, newUserLoad, "/company/1/adduser/")
	c.SetPath("/company/:company/adduser")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.AddUserToCompany(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Try to add the same user again
	c, rec = newContext(e, newUserLoad, "/company/1/adduser/")
	c.SetPath("/company/:company/adduser")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.AddUserToCompany(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Try to change access level for user
	newUserLoad, err = json.Marshal(models.UserPermissionRequest{UserID: 2, Access: models.Read})
	c, rec = newContext(e, newUserLoad, "/company/1/permission/")
	c.SetPath("/company/:company/permission")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.ChangeUserPermission(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Try to get company users
	c, rec = newContext(e, newUserLoad, "/company/1/users/")
	c.SetPath("/company/:company/users")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.GetCompanyUsers(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assertStruct(t, rec, []models.CompanyUserResponse{{UserID: 1, Username: "kristohb", Access: models.Admin}, {UserID: 2, Username: "doeman", Access: models.Read}})
	}

	// Try to delete john doe
	c, rec = newContext(e, newUserLoad, "/company/1/users/")
	c.SetPath("/company/:company/user")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.DeleteCompanyUser(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Check that john doe is truly removed
	c, rec = newContext(e, newUserLoad, "/company/1/users/")
	c.SetPath("/company/:company/users")
	c.SetParamNames("company")
	c.SetParamValues("1")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.GetCompanyUsers(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assertStruct(t, rec, []models.CompanyUserResponse{{UserID: 1, Username: "kristohb", Access: models.Admin}})
	}

	// Get companies for user
	c, rec = newContext(e, newUserLoad, "/user/companies/")
	c.SetPath("/user/companies")
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.GetUserCompanies(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assertStruct(t, rec, []models.Company{{ID: "1", Name: "coolio"}})
	}
}

func TestAddUser(t *testing.T) {

}
