package handler

import (
	"fmt"
	"net/http"

	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/utils"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"
)

// CreateCompany creates a new company
func (api RestAPI) CreateCompany(c echo.Context) (err error) {
	fmt.Println("create company:", c.Get(utils.UserIDContext.String()))
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return
	}

	var company models.Company
	if err = c.Bind(&company); err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "invalid body"})
	}

	if _, err = api.CompanyClient.CreateCompany(c.Request().Context(), company.Name, userID); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.createCompany: not able to save feedback", err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to create company"})
	}

	return c.JSON(http.StatusOK, web.HttpResponse{Message: "created company"})
}

func (api RestAPI) SearchCompany(c echo.Context) (err error) {
	query := c.Param("query")
	searchResult, err := api.CompanyClient.SearchCompany(c.Request().Context(), query)
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to perform query"})
	}

	return c.JSON(http.StatusOK, searchResult)
}
