package logger

import (
	"net/smtp"
	"strings"
)

type emailClient struct {
	from     string
	password string
	to       []string
	smtpHost string
	smtpPort string
}

func NewEmailClient(
	from string,
	password string,
	to []string,
	smtpHost string,
	smtpPort string,
) *emailClient {
	return &emailClient{
		from,
		password,
		to,
		smtpHost,
		smtpPort,
	}
}

func (e *emailClient) Email(subject string, body string) error {
	// This works for Gmail, not sure about others.
	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	message := "From: " + e.from + "\n" +
		"To: " + strings.Join(e.to, ",") + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	return smtp.SendMail(
		e.smtpHost+":"+e.smtpPort,
		auth,
		e.from,
		e.to,
		[]byte(message),
	)
}
