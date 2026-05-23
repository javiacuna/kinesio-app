package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Enabled   bool
	Host      string
	Port      string
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

type Message struct {
	ToEmail string
	Subject string
	Text    string
	HTML    string
}

type Mailer struct {
	cfg Config
}

func NewMailer(cfg Config) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(ctx context.Context, msg Message) error {
	if !m.cfg.Enabled {
		return nil
	}

	host := strings.TrimSpace(m.cfg.Host)
	port := strings.TrimSpace(m.cfg.Port)
	fromEmail := strings.TrimSpace(m.cfg.FromEmail)
	toEmail := strings.TrimSpace(msg.ToEmail)
	if host == "" || port == "" || fromEmail == "" {
		return errors.New("smtp_not_configured")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(fromEmail); err != nil {
		return err
	}

	address := net.JoinHostPort(host, port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}

	username := strings.TrimSpace(m.cfg.Username)
	password := strings.TrimSpace(m.cfg.Password)
	if username != "" || password != "" {
		if err := client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
			return err
		}
	}

	if err := client.Mail(fromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(m.buildMessage(msg)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (m *Mailer) buildMessage(msg Message) []byte {
	from := mail.Address{Name: strings.TrimSpace(m.cfg.FromName), Address: strings.TrimSpace(m.cfg.FromEmail)}
	to := mail.Address{Address: strings.TrimSpace(msg.ToEmail)}
	subject := strings.TrimSpace(msg.Subject)
	textBody := strings.TrimSpace(msg.Text)
	htmlBody := strings.TrimSpace(msg.HTML)
	if htmlBody == "" {
		htmlBody = "<p>" + html.EscapeString(textBody) + "</p>"
	}

	boundary := "kinesio-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	var buf bytes.Buffer
	writeHeader(&buf, "From", from.String())
	writeHeader(&buf, "To", to.String())
	writeHeader(&buf, "Subject", mime.QEncoding.Encode("UTF-8", subject))
	writeHeader(&buf, "MIME-Version", "1.0")
	writeHeader(&buf, "Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", boundary))
	buf.WriteString("\r\n")

	buf.WriteString("--" + boundary + "\r\n")
	writeHeader(&buf, "Content-Type", `text/plain; charset="UTF-8"`)
	writeHeader(&buf, "Content-Transfer-Encoding", "8bit")
	buf.WriteString("\r\n")
	buf.WriteString(textBody + "\r\n")

	buf.WriteString("--" + boundary + "\r\n")
	writeHeader(&buf, "Content-Type", `text/html; charset="UTF-8"`)
	writeHeader(&buf, "Content-Transfer-Encoding", "8bit")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody + "\r\n")
	buf.WriteString("--" + boundary + "--\r\n")
	return buf.Bytes()
}

func writeHeader(buf *bytes.Buffer, key string, value string) {
	buf.WriteString(key + ": " + value + "\r\n")
}
