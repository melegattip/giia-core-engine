package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type smtpEmailService struct {
	config *SMTPConfig
	logger pkgLogger.Logger
}

func NewSMTPEmailService(config *SMTPConfig, logger pkgLogger.Logger) providers.EmailService {
	return &smtpEmailService{
		config: config,
		logger: logger,
	}
}

func (s *smtpEmailService) SendActivationEmail(ctx context.Context, to, token, userName string) error {
	subject := "Activate Your Account"
	activationURL := fmt.Sprintf("https://yourapp.com/activate?token=%s", token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #007bff;
            color: #ffffff;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Welcome to GIIA, {{.UserName}}!</h2>
        <p>Thank you for registering. Please activate your account by clicking the button below:</p>
        <a href="{{.ActivationURL}}" class="button">Activate Account</a>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.ActivationURL}}</p>
        <p>This link will expire in 24 hours.</p>
        <div class="footer">
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"UserName":      userName,
		"ActivationURL": activationURL,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to render activation email template", nil)
		return err
	}

	return s.sendEmail(ctx, to, subject, body)
}

func (s *smtpEmailService) SendPasswordResetEmail(ctx context.Context, to, token, userName string) error {
	subject := "Reset Your Password"
	resetURL := fmt.Sprintf("https://yourapp.com/reset-password?token=%s", token)

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #dc3545;
            color: #ffffff;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Password Reset Request</h2>
        <p>Hi {{.UserName}},</p>
        <p>We received a request to reset your password. Click the button below to reset it:</p>
        <a href="{{.ResetURL}}" class="button">Reset Password</a>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">{{.ResetURL}}</p>
        <p>This link will expire in 1 hour.</p>
        <div class="footer">
            <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"UserName": userName,
		"ResetURL": resetURL,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to render password reset email template", nil)
		return err
	}

	return s.sendEmail(ctx, to, subject, body)
}

func (s *smtpEmailService) SendWelcomeEmail(ctx context.Context, to, userName string) error {
	subject := "Welcome to GIIA!"

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h2>Welcome to GIIA!</h2>
        <p>Hi {{.UserName}},</p>
        <p>Your account has been successfully activated. You can now log in and start using GIIA.</p>
        <p>If you have any questions, feel free to reach out to our support team.</p>
        <div class="footer">
            <p>Thank you for choosing GIIA!</p>
        </div>
    </div>
</body>
</html>
`

	data := map[string]string{
		"UserName": userName,
	}

	body, err := s.renderTemplate(tmpl, data)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to render welcome email template", nil)
		return err
	}

	return s.sendEmail(ctx, to, subject, body)
}

func (s *smtpEmailService) renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *smtpEmailService) sendEmail(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.config.From, to, subject, body)

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
	if err != nil {
		s.logger.Error(ctx, err, "Failed to send email", pkgLogger.Tags{
			"to":      to,
			"subject": subject,
		})
		return err
	}

	s.logger.Info(ctx, "Email sent successfully", pkgLogger.Tags{
		"to":      to,
		"subject": subject,
	})

	return nil
}
