package inits

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strconv"

	"github.com/rhaloubi/payment-gateway/auth-service/config"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	Dialer *gomail.Dialer
	From   string
}

func NewEmailService() *EmailService {
	port, err := strconv.Atoi(config.GetEnv("EMAIL_SMTP_PORT"))
	if err != nil {
		port = 587
	}

	return &EmailService{
		Dialer: gomail.NewDialer(
			config.GetEnv("EMAIL_SMTP_HOST"),
			port,
			config.GetEnv("EMAIL_SMTP_USER"),
			config.GetEnv("EMAIL_SMTP_PASS"),
		),
		From: config.GetEnv("EMAIL_FROM"),
	}
}

func (e *EmailService) RenderTemplate(file string, data interface{}) (string, error) {
	// adjust the path relative to your project structure
	templatePath := filepath.Join("internal", "emails", "templates", file)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (e *EmailService) SendHTML(to, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	return e.Dialer.DialAndSend(m)
}
