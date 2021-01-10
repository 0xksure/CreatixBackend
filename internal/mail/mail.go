package mail

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type MailClienter interface {
	SendEmail(from, to, content string) error
}

type MailClient struct {
	SendgridClient *sendgrid.Client
}

type NewMail struct {
	Email   string `json:"email"`
	Content string `json:"content"`
}

func NewMailClient(sendgridKey string) *MailClient {
	client := sendgrid.NewSendClient(sendgridKey)
	return &MailClient{SendgridClient: client}
}

func (c *MailClient) SendEmail(fromAddress, fromName, toAddress, toName, subject, content string) (statusCode int, err error) {
	from := mail.NewEmail(fromName, fromAddress)
	to := mail.NewEmail(toName, toAddress)

	message := mail.NewSingleEmail(from, subject, to, content, "")

	resp, err := c.SendgridClient.Send(message)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode > 300 {
		return resp.StatusCode, err
	}
	return resp.StatusCode, err
}
