package handler

import (
	"net/http"

	"github.com/kristohberg/CreatixBackend/config"
	"github.com/kristohberg/CreatixBackend/internal/mail"
	"github.com/kristohberg/CreatixBackend/web"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type PublicAPI struct {
	Cfg        config.Config
	Logging    *logrus.Logger
	MailClient *mail.MailClient
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
	var newMail mail.NewMail
	err := c.Bind(&newMail)
	if err != nil {
		p.Logging.WithError(err)
		return c.JSON(http.StatusBadRequest, web.HttpResponse{Message: "not able to bind contact content"})
	}

	statusCode, err := p.MailClient.SendEmail(newMail.Email, "Contact Us", p.Cfg.ContactEmail, "TheCreatix", "Creatix: Contact us", newMail.Content)
	if err != nil {
		p.Logging.Fatalf("error when sending email: %v", err)
		return c.String(http.StatusBadRequest, "")
	}
	return c.JSON(statusCode, "")

}
