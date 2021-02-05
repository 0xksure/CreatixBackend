package handler

import (
	"net/http"

	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var (
	AddNewUserByEmailToCompanyPath = "/company/:company/adduser"
	POSTCreateNewCompanyPath       = "/company/create"
)

func (api RestAPI) CompanyHandler(e *echo.Group) {
	e.GET("/company/search/{query}", api.SearchCompany)

	e.POST(POSTCreateNewCompanyPath, api.CreateCompany)
	e.POST(AddNewUserByEmailToCompanyPath, api.AddUserByEmailToCompany)
	e.POST("/company/:company/permission", api.ChangeUserPermission)
	e.GET("/company/:company/users", api.GetCompanyUsers)
	e.GET("/user/companies", api.GetUserCompanies)
	e.DELETE("/company/:company/user/:userid", api.DeleteCompanyUser)
}

// CreateCompany creates a new company
func (api RestAPI) CreateCompany(c echo.Context) (err error) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		api.Logging.Unsuccessful("creatix.feedback.createCompany: could not get user", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "no userID"})
	}

	if userID == "" {
		api.Logging.Unsuccessful("creatix.feedback.createCompany: could not get user", nil)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "no userID"})
	}

	var company models.Company
	if err = c.Bind(&company); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.createCompany: could not bind company", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "invalid body"})
	}

	if _, err = api.CompanyClient.CreateCompany(c.Request().Context(), company.Name, userID); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.createCompany: not able to create company", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to create company"})
	}

	return c.JSON(http.StatusOK, web.HttpResponse{Message: "created company"})
}

func (api RestAPI) AddUserByEmailToCompany(c echo.Context) (err error) {
	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("not authorized", errors.New("userID not provided"))
		return c.String(http.StatusUnauthorized, "could not get userid")
	}

	err = api.SessionClient.IsAuthorized(c.Request().Context(), userID, companyID, models.Admin)
	if err != nil {
		api.Logging.Unsuccessful("not authorized", err)
		return c.String(http.StatusUnauthorized, "")
	}

	newUserRequest := new(models.AddUser)
	err = c.Bind(newUserRequest)
	if err != nil {
		api.Logging.Unsuccessful("could not bind user", err)
		return c.String(http.StatusBadRequest, "")
	}

	if newUserRequest.Username != "" {
		err = api.CompanyClient.AddUserToCompanyByUsername(c.Request().Context(), companyID, *newUserRequest)
		if err != nil {
			api.Logging.Unsuccessful("could not add user", err)
			return c.String(http.StatusBadRequest, "")
		}
	} else if newUserRequest.Email != "" {
		err = api.CompanyClient.AddUserToCompanyByEmail(c.Request().Context(), companyID, *newUserRequest)
		if err != nil {
			api.Logging.Unsuccessful("could not add user", err)
			return c.String(http.StatusBadRequest, "")
		}
	} else {
		api.Logging.Unsuccessful("could not add user", errors.New("no user specified"))
		return c.String(http.StatusBadRequest, "")
	}

	return c.String(http.StatusOK, companyID)
}

func (api RestAPI) DeleteCompanyUser(c echo.Context) (err error) {
	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		return errors.WithStack(errors.New("could not get user id"))
	}

	err = api.SessionClient.IsAuthorized(c.Request().Context(), userID, companyID, models.Admin)
	if err != nil {
		api.Logging.Unsuccessful("not authorized  ", err)
		return c.String(http.StatusUnauthorized, "")
	}

	userIDToDelete := c.Param("userid")
	if userIDToDelete == "" {
		api.Logging.Unsuccessful("no userid to delete provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	err = api.CompanyClient.DeleteUser(c.Request().Context(), companyID, userIDToDelete)
	if err != nil {
		api.Logging.Unsuccessful("could not delete user", err)
		return c.String(http.StatusInternalServerError, "")
	}

	return c.String(http.StatusOK, "ok")

}

func (api RestAPI) GetCompanyUsers(c echo.Context) (err error) {
	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		return errors.WithStack(errors.New("could not get user id"))
	}

	err = api.SessionClient.IsAuthorized(c.Request().Context(), userID, companyID, models.Admin)
	if err != nil {
		api.Logging.Unsuccessful("not authorized  ", err)
		return c.String(http.StatusUnauthorized, "")
	}

	companyUsers, err := api.CompanyClient.GetCompanyUsers(c.Request().Context(), companyID)
	if err != nil {
		api.Logging.Unsuccessful("not able to get company users  ", err)
		return c.String(http.StatusInternalServerError, "")
	}
	return c.JSON(http.StatusOK, companyUsers)
}

func (api RestAPI) GetUserCompanies(c echo.Context) (err error) {
	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		api.Logging.Unsuccessful("could not get user id", err)
		return c.String(http.StatusBadRequest, "")
	}

	companies, err := api.CompanyClient.GetUserCompanies(c.Request().Context(), userID)
	if err != nil {
		api.Logging.Unsuccessful("not able to get companies for user", err)
		return c.String(http.StatusBadRequest, "")
	}

	return c.JSON(http.StatusOK, companies)
}

func (api RestAPI) ChangeUserPermission(c echo.Context) (err error) {
	companyID := c.Param("company")
	if companyID == "" {
		api.Logging.Unsuccessful("no company provided ", nil)
		return c.String(http.StatusBadRequest, "")
	}

	userID := c.Get(utils.UserIDContext.String()).(string)
	if userID == "" {
		return errors.WithStack(errors.New("could not get user id"))
	}

	err = api.SessionClient.IsAuthorized(c.Request().Context(), userID, companyID, models.Admin)
	if err != nil {
		api.Logging.Unsuccessful("not authorized  ", err)
		return c.String(http.StatusUnauthorized, "")
	}

	newUserRequest := new(models.UserPermissionRequest)
	err = c.Bind(newUserRequest)
	if err != nil {
		api.Logging.Unsuccessful("could not bind user", err)
		return c.String(http.StatusBadRequest, "")
	}
	err = api.CompanyClient.UpdateUserPermission(c.Request().Context(), companyID, *newUserRequest)
	if err != nil {
		api.Logging.Unsuccessful("could not add user", err)
		return c.String(http.StatusBadRequest, "")
	}
	return c.String(http.StatusOK, companyID)

}
func (api RestAPI) SearchCompany(c echo.Context) (err error) {
	query := c.Param("query")
	searchResult, err := api.CompanyClient.SearchCompany(c.Request().Context(), query)
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to perform query"})
	}

	return c.JSON(http.StatusOK, searchResult)
}
