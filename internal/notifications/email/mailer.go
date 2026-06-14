package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
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

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Message struct {
	ToEmail     string
	ReplyTo     string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
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

	now := time.Now().UnixNano()
	altBoundary := "kinesio-alt-" + strconv.FormatInt(now, 10)
	var buf bytes.Buffer
	writeHeader(&buf, "From", from.String())
	writeHeader(&buf, "To", to.String())
	if replyTo := strings.TrimSpace(msg.ReplyTo); replyTo != "" {
		if _, err := mail.ParseAddress(replyTo); err == nil {
			writeHeader(&buf, "Reply-To", replyTo)
		}
	}
	writeHeader(&buf, "Subject", mime.QEncoding.Encode("UTF-8", subject))
	writeHeader(&buf, "MIME-Version", "1.0")

	hasAttachments := len(msg.Attachments) > 0
	mixedBoundary := "kinesio-mixed-" + strconv.FormatInt(now, 10)
	if hasAttachments {
		writeHeader(&buf, "Content-Type", fmt.Sprintf("multipart/mixed; boundary=%q", mixedBoundary))
		buf.WriteString("\r\n")
		buf.WriteString("--" + mixedBoundary + "\r\n")
	}

	writeHeader(&buf, "Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", altBoundary))
	buf.WriteString("\r\n")

	buf.WriteString("--" + altBoundary + "\r\n")
	writeHeader(&buf, "Content-Type", `text/plain; charset="UTF-8"`)
	writeHeader(&buf, "Content-Transfer-Encoding", "8bit")
	buf.WriteString("\r\n")
	buf.WriteString(textBody + "\r\n")

	buf.WriteString("--" + altBoundary + "\r\n")
	writeHeader(&buf, "Content-Type", `text/html; charset="UTF-8"`)
	writeHeader(&buf, "Content-Transfer-Encoding", "8bit")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody + "\r\n")
	buf.WriteString("--" + altBoundary + "--\r\n")

	if hasAttachments {
		for _, attachment := range msg.Attachments {
			contentType := strings.TrimSpace(attachment.ContentType)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			filename := strings.TrimSpace(attachment.Filename)
			if filename == "" {
				filename = "attachment"
			}

			buf.WriteString("--" + mixedBoundary + "\r\n")
			writeHeader(&buf, "Content-Type", fmt.Sprintf("%s; name=%q", contentType, filename))
			writeHeader(&buf, "Content-Transfer-Encoding", "base64")
			writeHeader(&buf, "Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
			buf.WriteString("\r\n")
			writeBase64(&buf, attachment.Data)
		}
		buf.WriteString("--" + mixedBoundary + "--\r\n")
	}

	return buf.Bytes()
}

func writeBase64(buf *bytes.Buffer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLength = 76
	for len(encoded) > 0 {
		n := min(lineLength, len(encoded))
		buf.WriteString(encoded[:n])
		buf.WriteString("\r\n")
		encoded = encoded[n:]
	}
}

func writeHeader(buf *bytes.Buffer, key string, value string) {
	buf.WriteString(key + ": " + value + "\r\n")
}
