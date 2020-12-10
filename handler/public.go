package handler

import (
	"net/http"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/sirupsen/logrus"
)

type ContactUsContent struct {
	Email   string `json:"email"`
	Content string `json:"content"`
}

type PublicAPI struct {
	Cfg     config.Config
	Logging *logrus.Logger
}

func (p PublicAPI) Handler(e *echo.Group) {
	e.GET("/health", p.Health)
	e.POST("/contact-us", p.ContactUs)
}

func (p PublicAPI) Health(c echo.Context) error {
	return c.String(http.StatusOK, "I'm up")
}

// ContactUs handles
func (p PublicAPI) ContactUs(c echo.Context) error {
	var contactUsContent ContactUsContent
	err := c.Bind(&contactUsContent)
	if err != nil {
		p.Logging.WithError(err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to bind contact content"})
	}
	from := mail.NewEmail("ContactUs", contactUsContent.Email)
	subject := "creatix: Contact us"
	to := mail.NewEmail("thecreatix", p.Cfg.ContactEmail)
	plainTextContent := contactUsContent.Content
	htmlContent := ""
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(p.Cfg.SendgridKey)
	response, err := client.Send(message)
	if err != nil {
		p.Logging.WithError(err)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	if response.StatusCode > 300 {
		p.Logging.WithError(err)
		return c.JSON(response.StatusCode, nil)
	}
	return c.JSON(response.StatusCode, "")

}
