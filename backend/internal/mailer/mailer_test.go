package mailer

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestFormatAddress(t *testing.T) {
	if got := formatAddress("mail@example.com", ""); got != "mail@example.com" {
		t.Fatalf("unexpected address without name: %q", got)
	}
	if got := formatAddress("mail@example.com", "  Forstitch  "); got != "Forstitch <mail@example.com>" {
		t.Fatalf("unexpected named address: %q", got)
	}
}

func TestNoopMailer(t *testing.T) {
	if err := (Noop{}).Send("mail@example.com", "Subject", "Body"); err != nil {
		t.Fatalf("noop mailer returned error: %v", err)
	}
}

func TestSMTPSend(t *testing.T) {
	host, port, messages, stop := startSMTPTestServer(t)
	defer stop()

	sender := SMTP{
		Host: host, Port: port, Username: "sender@example.com", Password: "secret", FromName: "Forstitch",
	}
	if err := sender.Send("buyer@example.com", "Код подтверждения", "Код: 123456"); err != nil {
		t.Fatalf("send mail: %v", err)
	}

	select {
	case message := <-messages:
		if message.from != "sender@example.com" || message.to != "buyer@example.com" {
			t.Fatalf("unexpected SMTP envelope: %+v", message)
		}
		if message.username != "sender@example.com" || message.password != "secret" {
			t.Fatalf("unexpected SMTP auth: %+v", message)
		}
		expectedParts := []string{
			"From: Forstitch <sender@example.com>",
			"To: buyer@example.com",
			"Subject: Код подтверждения",
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
			"Код: 123456",
		}
		for _, part := range expectedParts {
			if !strings.Contains(message.data, part) {
				t.Fatalf("expected message to contain %q, got:\n%s", part, message.data)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SMTP message")
	}
}

func TestSMTPSendConnectionError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	address := listener.Addr().(*net.TCPAddr)
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved port: %v", err)
	}

	sender := SMTP{Host: "127.0.0.1", Port: fmt.Sprintf("%d", address.Port), From: "sender@example.com"}
	if err := sender.Send("buyer@example.com", "Subject", "Body"); err == nil {
		t.Fatal("expected SMTP connection error")
	}
}

type smtpTestMessage struct {
	from     string
	to       string
	username string
	password string
	data     string
}

func startSMTPTestServer(t *testing.T) (string, string, <-chan smtpTestMessage, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen SMTP: %v", err)
	}
	result := make(chan smtpTestMessage, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()

		reader := bufio.NewReader(connection)
		writer := bufio.NewWriter(connection)
		writeSMTPLine(writer, "220 localhost ESMTP test")
		var message smtpTestMessage
		var data strings.Builder
		inData := false

		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					message.data = data.String()
					writeSMTPLine(writer, "250 queued")
					continue
				}
				data.WriteString(line)
				data.WriteString("\n")
				continue
			}

			switch {
			case strings.HasPrefix(line, "EHLO "), strings.HasPrefix(line, "HELO "):
				_, _ = writer.WriteString("250-localhost\r\n250 AUTH PLAIN\r\n")
				_ = writer.Flush()
			case strings.HasPrefix(line, "AUTH PLAIN"):
				parts := strings.Fields(line)
				if len(parts) == 3 {
					decoded, decodeErr := base64.StdEncoding.DecodeString(parts[2])
					if decodeErr == nil {
						authParts := strings.Split(string(decoded), "\x00")
						if len(authParts) == 3 {
							message.username = authParts[1]
							message.password = authParts[2]
						}
					}
				}
				writeSMTPLine(writer, "235 authenticated")
			case strings.HasPrefix(line, "MAIL FROM:"):
				message.from = smtpAddress(line)
				writeSMTPLine(writer, "250 sender ok")
			case strings.HasPrefix(line, "RCPT TO:"):
				message.to = smtpAddress(line)
				writeSMTPLine(writer, "250 recipient ok")
			case line == "DATA":
				inData = true
				writeSMTPLine(writer, "354 end with dot")
			case line == "QUIT":
				result <- message
				writeSMTPLine(writer, "221 bye")
				return
			default:
				writeSMTPLine(writer, "250 ok")
			}
		}
	}()

	address := listener.Addr().(*net.TCPAddr)
	stop := func() {
		_ = listener.Close()
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
	return "127.0.0.1", fmt.Sprintf("%d", address.Port), result, stop
}

func writeSMTPLine(writer *bufio.Writer, line string) {
	_, _ = writer.WriteString(line + "\r\n")
	_ = writer.Flush()
}

func smtpAddress(line string) string {
	start := strings.IndexByte(line, '<')
	end := strings.LastIndexByte(line, '>')
	if start >= 0 && end > start {
		return line[start+1 : end]
	}
	return ""
}
