package mailer

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SMTP struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string
}

func (m SMTP) Send(to string, subject string, body string) error {
	addr := net.JoinHostPort(m.Host, m.Port)
	from := strings.TrimSpace(m.From)
	if from == "" {
		from = m.Username
	}

	headers := map[string]string{
		"From":         formatAddress(from, m.FromName),
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var message strings.Builder
	for key, value := range headers {
		message.WriteString(key)
		message.WriteString(": ")
		message.WriteString(value)
		message.WriteString("\r\n")
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	var auth smtp.Auth
	if m.Username != "" || m.Password != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(message.String()))
}

func formatAddress(email string, name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
