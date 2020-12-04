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
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignupUser(t *testing.T) {
	// Setup
	e := echo.New()

	signupLoad := models.Signup{User: models.User{Firstname: "Kris", Lastname: "Berg", Email: "ok@ok.com", Password: "olol"}}
	signupJSON, err := json.Marshal(signupLoad)

	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(signupJSON)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	db, err := test.NewTestDB()
	defer test.EmptyTestDB(db)

	require.NoError(t, err)
	sessionAPI := SessionAPI{
		DB:          db,
		Logging:     logging.NewLogger(),
		Cfg:         config.Config{},
		UserSession: models.UserSession{},
	}

	// Signup user
	res := sessionAPI.Signup(c)
	if assert.NoError(t, res) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, string("user created"), rec.Body.String())
	}

	// Try to signup the same user again
	res = sessionAPI.Signup(c)
	if assert.NoError(t, res) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, string("error"), rec.Body.String())
	}
}
