package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/logging"
	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/test"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newContext(e *echo.Echo, data []byte, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(data)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec

}

func newUser() models.User {
	return models.User{ID: "1", Firstname: "Kris", Lastname: "Berg", Username: "kristohb", Email: "ok@ok.com", Password: "olol"}
}

func newCompany() models.Company {
	return models.Company{Name: "coolio"}
}

func newSessionUser(user models.User) utils.SessionUser {
	return utils.SessionUser{ID: user.ID, Firstname: user.Firstname, Lastname: user.Lastname, Email: user.Email}
}

func newLoginRequest(user models.User) models.LoginRequest {
	return models.LoginRequest{Email: user.Email, Password: user.Password}
}

func getStringJSON(rec *httptest.ResponseRecorder) string {
	return strings.TrimSuffix(rec.Body.String(), "\n")
}

func assertStruct(t *testing.T, received *httptest.ResponseRecorder, expected interface{}) bool {
	expectedResp, err := json.Marshal(expected)
	require.NoError(t, err)
	return assert.Equal(t, string(expectedResp), getStringJSON(received))
}

func TestSession(t *testing.T) {
	// Setup
	var c echo.Context
	var rec *httptest.ResponseRecorder
	e := echo.New()

	db, err := test.NewTestDB()
	require.NoError(t, err)
	defer test.EmptyTestDB(t, db)

	logger := logging.NewLogger()
	sessionAPI := SessionAPI{
		DB:            db,
		Logging:       logger,
		Cfg:           config.Config{Env: "test"},
		SessionClient: models.NewSessionClient(db, []byte("secret"), 20, logger),
	}

	// Signup user
	mockUser := newUser()
	signupLoad := models.Signup{User: mockUser,
		Company: newCompany()}
	signupJSON, err := json.Marshal(signupLoad)
	require.NoError(t, err)
	c, rec = newContext(e, signupJSON, "/")
	res := sessionAPI.Signup(c)
	if assert.NoError(t, res) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, string("user created"), rec.Body.String())
	}

	// Try to signup the same user again
	c, rec = newContext(e, signupJSON, "/")
	res = sessionAPI.Signup(c)

	if assert.NoError(t, res) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, string("could not create user"), rec.Body.String())
	}

	// try to login as user
	loginRequest := models.LoginRequest{Email: "ok@ok.com", Password: "olol"}
	loginRequestByte, err := json.Marshal(loginRequest)
	require.NoError(t, err)
	c, rec = newContext(e, loginRequestByte, "/")
	err = sessionAPI.Login(c)

	require.NoError(t, err)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assertStruct(t, rec, newSessionUser(mockUser))
	}
	cookie := rec.Header().Get("Set-Cookie")

	// Try to refresh cookie
	t.Log("Test to refresh cookie")
	c, rec = newContext(e, nil, "/")
	c.Set("cookie", cookie)
	err = sessionAPI.Refresh(c)
	if assert.NoError(t, err) {
		newCookie := rec.Header().Get("Set-Cookie")
		assert.NotEqual(t, newCookie, cookie)
	}

	// Try to login with nonexisting user
	t.Log("Login with nonexisting user")
	loginRequest = models.LoginRequest{Email: "ok2@ok.com", Password: "olol"}
	loginRequestByte, err = json.Marshal(loginRequest)
	require.NoError(t, err)
	c, rec = newContext(e, loginRequestByte, "/")
	err = sessionAPI.Login(c)

	require.NoError(t, err)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

	// Try to logout
	t.Log("Test to logout cookie")
	c, rec = newContext(e, nil, "/")
	c.Set("cookie", cookie)
	err = sessionAPI.Logout(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		newCookie := rec.Header().Get("Set-Cookie")
		assert.NotEqual(t, newCookie, cookie)
	}

	// Try to refresh without cookie
	t.Log("Try to refresh without cookie")
	c, rec = newContext(e, nil, "/")
	err = sessionAPI.Refresh(c)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}

}
