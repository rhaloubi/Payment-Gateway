package service

import (
	"crypto/tls"
	"fmt"
	"os"

	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	frontendURL  string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     getEnv("MAILTRAP_HOST", "sandbox.smtp.mailtrap.io"),
		smtpPort:     getEnvInt("MAILTRAP_PORT", 2525),
		smtpUsername: os.Getenv("MAILTRAP_USERNAME"),
		smtpPassword: os.Getenv("MAILTRAP_PASSWORD"),
		fromEmail:    getEnv("FROM_EMAIL", "noreply@paymentgateway.ma"),
		fromName:     getEnv("FROM_NAME", "Payment Gateway Morocco"),
		frontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(invitation *model.MerchantInvitation, merchant *model.Merchant) error {
	// Build invitation URL
	invitationURL := fmt.Sprintf("%s/invitations/accept/%s", s.frontendURL, invitation.InvitationToken)

	// Email subject
	subject := fmt.Sprintf("You've been invited to join %s", merchant.BusinessName)

	// Email body (HTML)
	body := s.buildInvitationEmailHTML(merchant.BusinessName, invitationURL, invitation.ExpiresAt.Format("January 2, 2006"))

	// Send email
	return s.sendEmail(invitation.Email, subject, body)
}

// buildInvitationEmailHTML builds the HTML email template
func (s *EmailService) buildInvitationEmailHTML(merchantName, invitationURL, expiresAt string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; }
        .button { display: inline-block; background-color: #4F46E5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Team Invitation</h1>
        </div>
        <div class="content">
            <h2>You've been invited!</h2>
            <p>You have been invited to join <strong>%s</strong> on Payment Gateway Morocco.</p>
            <p>Click the button below to accept the invitation and join the team:</p>
            <center>
                <a href="%s" class="button">Accept Invitation</a>
            </center>
            <p style="margin-top: 30px; font-size: 14px; color: #6b7280;">
                This invitation will expire on <strong>%s</strong>.
            </p>
            <p style="margin-top: 20px; font-size: 14px; color: #6b7280;">
                If the button doesn't work, copy and paste this link into your browser:<br>
                <a href="%s">%s</a>
            </p>
        </div>
        <div class="footer">
            <p>© 2025 Payment Gateway Morocco. All rights reserved.</p>
            <p>This is an automated email. Please do not reply.</p>
        </div>
    </div>
</body>
</html>
	`, merchantName, invitationURL, expiresAt, invitationURL, invitationURL)
}

// sendEmail sends an email via Mailtrap
func (s *EmailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.smtpUsername, s.smtpPassword)

	// For Mailtrap, we can use TLS
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
