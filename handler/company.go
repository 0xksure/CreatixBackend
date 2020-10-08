package handler

import (
	"net/http"

	"github.com/kristohberg/CreatixBackend/models"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"
)

func (api RestAPI) SearchCompany(c echo.Context) (err error) {
	query := c.Param("query")
	searchResult, err := api.CompanyAPI.SearchCompany(c.Request().Context(), api.DB, query)
	if err != nil {
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to perform query"})
	}

	return c.JSON(http.StatusOK, searchResult)
}

// PostFeedback posts feedback from user
func (api RestAPI) CreateCompany(c echo.Context) (err error) {
	var company models.Company
	if err = c.Bind(&company); err != nil {
		return
	}
	if _, err = company.CreateCompany(c.Request().Context(), api.DB); err != nil {
		api.Logging.Unsuccessful("creatix.feedback.postfeedback: not able to save feedback", err)
		return
	}
	return c.JSON(http.StatusOK, web.HttpResponse{Message: "created company"})
}
