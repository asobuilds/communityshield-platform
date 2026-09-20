package services

import (
	"errors"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// Send delivers an HTML email via SMTP. All config comes from env:
//   SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
// If any required var is missing, returns an error (does not crash).
// Fail-closed: an unconfigured server never silently "succeeds".
func (s *EmailService) Send(to, subject, htmlBody string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := os.Getenv("SMTP_PASS")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if host == "" || port == "" || from == "" {
		return errors.New("email service is not configured (SMTP_HOST/PORT/FROM missing)")
	}
	if strings.TrimSpace(to) == "" {
		return errors.New("recipient email is required")
	}

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return smtp.SendMail(addr, auth, extractEmail(from), []string{to}, []byte(msg.String()))
}

// extractEmail pulls the raw address out of a "Name <addr@host>" form.
func extractEmail(from string) string {
	if i := strings.Index(from, "<"); i >= 0 {
		if j := strings.Index(from[i:], ">"); j > 0 {
			return from[i+1 : i+j]
		}
	}
	return from
}