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

	mockUser := newUser()
	signupLoad := models.Signup{User: mockUser, Company: newCompany()}
	signupJSON, err := json.Marshal(signupLoad)
	require.NoError(t, err)
	c, rec = newContext(e, signupJSON)
	// SIGNUP USER
	res := sessionAPI.Signup(c)
	require.NoError(t, res)

	// Login user
	loginRequestByte, err := json.Marshal(newLoginRequest(mockUser))
	require.NoError(t, err)
	c, rec = newContext(e, loginRequestByte)
	err = sessionAPI.Login(c)
	require.NoError(t, err)

	// get cookie

	return rec.Header().Get("Set-Cookie")
}

func NewRestAPI(db *sql.DB, logger *logging.StandardLogger) RestAPI {
	cfg := InitConfig()
	return RestAPI{
		DB:            db,
		Logging:       logger,
		Cfg:           cfg,
		Feedback:      models.Feedback{},
		Middleware:    &middleware.Middleware{Cfg: cfg},
		CompanyClient: models.NewCompanyClient(db),
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
	defer test.EmptyTestDB(db)

	// Create user
	logger := logging.NewLogger()
	restAPI := NewRestAPI(db, logger)

	// Create new company
	companyJSON, err := json.Marshal(newCompany())
	require.NoError(t, err)
	c, rec = newContext(e, companyJSON)
	c.Set(utils.UserIDContext.String(), "1")
	err = restAPI.CreateCompany(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}

}
