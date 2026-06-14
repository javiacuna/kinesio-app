package http

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/notifications/email"
	"github.com/javiacuna/kinesio-backend/internal/support/domain"
)

const maxAttachmentBytes = 10 << 20

type Mailer interface {
	Send(ctx context.Context, msg email.Message) error
}

type repository interface {
	Create(ctx context.Context, ticket domain.Ticket) (domain.Ticket, error)
	ListByUser(ctx context.Context, email string, limit int) ([]domain.Ticket, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Ticket, bool, error)
}

type Handler struct {
	repo         repository
	mailer       Mailer
	supportEmail string
	supportPhone string
	rootDir      string
}

func NewHandler(repo repository, mailer Mailer, supportEmail string, supportPhone string, rootDir string) *Handler {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		rootDir = "uploads/support-attachments"
	}
	return &Handler{
		repo:         repo,
		mailer:       mailer,
		supportEmail: strings.TrimSpace(supportEmail),
		supportPhone: strings.TrimSpace(supportPhone),
		rootDir:      filepath.Clean(rootDir),
	}
}

type infoResponse struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (h *Handler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, infoResponse{
		Email: h.supportEmail,
		Phone: h.supportPhone,
	})
}

type ticketResponse struct {
	ID                 string `json:"id"`
	Subject            string `json:"subject"`
	Message            string `json:"message"`
	AttachmentFileName string `json:"attachment_file_name,omitempty"`
	AttachmentURL      string `json:"attachment_url,omitempty"`
	CreatedAt          string `json:"created_at"`
}

func (h *Handler) Contact(c *gin.Context) {
	subject := strings.TrimSpace(c.PostForm("subject"))
	message := strings.TrimSpace(c.PostForm("message"))
	if subject == "" || message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject_and_message_required"})
		return
	}

	userEmail := ""
	userRole := ""
	if user, ok := middleware.CurrentUser(c); ok {
		userEmail = user.Email
		userRole = user.Role
	}

	id := uuid.New()
	var attachments []email.Attachment
	var attachmentFileName, attachmentContentType, attachmentStoragePath *string
	var attachmentSizeBytes *int64

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachmentBytes)
	if file, header, err := c.Request.FormFile("file"); err == nil {
		defer file.Close()
		if header.Size > maxAttachmentBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large"})
			return
		}
		data, err := io.ReadAll(io.LimitReader(file, maxAttachmentBytes+1))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if len(data) > maxAttachmentBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large"})
			return
		}
		contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if err := os.MkdirAll(h.rootDir, 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		storagePath := filepath.Join(h.rootDir, id.String()+strings.ToLower(filepath.Ext(header.Filename)))
		if err := os.WriteFile(storagePath, data, 0o644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}

		fileName := filepath.Base(header.Filename)
		size := int64(len(data))
		attachmentFileName = &fileName
		attachmentContentType = &contentType
		attachmentSizeBytes = &size
		attachmentStoragePath = &storagePath

		attachments = append(attachments, email.Attachment{
			Filename:    fileName,
			ContentType: contentType,
			Data:        data,
		})
	}

	var userRolePtr *string
	if userRole != "" {
		userRolePtr = &userRole
	}

	ticket := domain.Ticket{
		ID:                    id,
		UserEmail:             userEmail,
		UserRole:              userRolePtr,
		Subject:               subject,
		Message:               message,
		AttachmentFileName:    attachmentFileName,
		AttachmentContentType: attachmentContentType,
		AttachmentSizeBytes:   attachmentSizeBytes,
		AttachmentStoragePath: attachmentStoragePath,
	}

	created, err := h.repo.Create(c.Request.Context(), ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	if h.mailer != nil && h.supportEmail != "" {
		msg := email.Message{
			ToEmail:     h.supportEmail,
			ReplyTo:     userEmail,
			Subject:     "[Soporte Kinesio App] " + subject,
			Text:        contactText(userEmail, userRole, subject, message),
			HTML:        contactHTML(userEmail, userRole, subject, message),
			Attachments: attachments,
		}
		if err := h.mailer.Send(c.Request.Context(), msg); err != nil {
			c.JSON(http.StatusOK, gin.H{"ok": true, "email_sent": false})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "email_sent": true, "ticket": toResponse(created)})
}

func (h *Handler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tickets, err := h.repo.ListByUser(c.Request.Context(), user.Email, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]ticketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, toResponse(ticket))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) DownloadAttachment(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("ticket_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_ticket_id"})
		return
	}

	ticket, found, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found || !strings.EqualFold(ticket.UserEmail, user.Email) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if ticket.AttachmentStoragePath == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if _, err := os.Stat(*ticket.AttachmentStoragePath); errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file_not_found"})
		return
	}

	fileName := "attachment"
	if ticket.AttachmentFileName != nil {
		fileName = *ticket.AttachmentFileName
	}
	if ticket.AttachmentContentType != nil {
		c.Header("Content-Type", *ticket.AttachmentContentType)
	}
	c.FileAttachment(*ticket.AttachmentStoragePath, fileName)
}

func toResponse(ticket domain.Ticket) ticketResponse {
	resp := ticketResponse{
		ID:        ticket.ID.String(),
		Subject:   ticket.Subject,
		Message:   ticket.Message,
		CreatedAt: ticket.CreatedAt.UTC().Format(time.RFC3339),
	}
	if ticket.AttachmentFileName != nil {
		resp.AttachmentFileName = *ticket.AttachmentFileName
		resp.AttachmentURL = "/api/v1/support/tickets/" + ticket.ID.String() + "/attachment"
	}
	return resp
}

func contactText(userEmail, userRole, subject, message string) string {
	var b strings.Builder
	b.WriteString("Nuevo reclamo/consulta de soporte\n\n")
	if userEmail != "" {
		fmt.Fprintf(&b, "Usuario: %s (%s)\n", userEmail, userRole)
	}
	fmt.Fprintf(&b, "Asunto: %s\n\n", subject)
	b.WriteString(message)
	return b.String()
}

func contactHTML(userEmail, userRole, subject, message string) string {
	var userLine string
	if userEmail != "" {
		userLine = fmt.Sprintf(`<p style="margin:0 0 8px"><strong>Usuario:</strong> %s (%s)</p>`, html.EscapeString(userEmail), html.EscapeString(userRole))
	}
	return `<!doctype html>
<html>
  <body style="font-family:Arial,sans-serif;color:#111827;line-height:1.5">
    <h2 style="margin:0 0 12px">Nuevo reclamo/consulta de soporte</h2>
    ` + userLine + `
    <p style="margin:0 0 8px"><strong>Asunto:</strong> ` + html.EscapeString(subject) + `</p>
    <p style="margin:0 0 16px;white-space:pre-wrap">` + html.EscapeString(message) + `</p>
    <p style="margin:0;color:#6b7280;font-size:13px">Kinesio App</p>
  </body>
</html>`
}
