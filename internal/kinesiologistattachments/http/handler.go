package http

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologistattachments/domain"
)

const maxUploadBytes = 50 << 20

type repository interface {
	Create(ctx context.Context, attachment domain.Attachment) (domain.Attachment, error)
	ListByKinesiologist(ctx context.Context, kinesiologistID uuid.UUID) ([]domain.Attachment, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Attachment, bool, error)
}

type Handler struct {
	repo    repository
	rootDir string
}

func NewHandler(repo repository, rootDir string) *Handler {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		rootDir = "uploads/kinesiologist-attachments"
	}
	return &Handler{repo: repo, rootDir: filepath.Clean(rootDir)}
}

type attachmentResponse struct {
	ID              string  `json:"id"`
	KinesiologistID string  `json:"kinesiologist_id"`
	FileName        string  `json:"file_name"`
	ContentType     string  `json:"content_type"`
	SizeBytes       int64   `json:"size_bytes"`
	Kind            string  `json:"kind"`
	Category        string  `json:"category"`
	Notes           *string `json:"notes,omitempty"`
	UploadedByEmail *string `json:"uploaded_by_email,omitempty"`
	UploadedByRole  *string `json:"uploaded_by_role,omitempty"`
	DownloadURL     string  `json:"download_url"`
	CreatedAt       string  `json:"created_at"`
}

func (h *Handler) ListByKinesiologist(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	kinesiologistID, err := uuid.Parse(strings.TrimSpace(c.Param("kinesiologist_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_kinesiologist_id"})
		return
	}

	items, err := h.repo.ListByKinesiologist(c.Request.Context(), kinesiologistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]attachmentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) Upload(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	kinesiologistID, err := uuid.Parse(strings.TrimSpace(c.Param("kinesiologist_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_kinesiologist_id"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_required"})
		return
	}
	defer file.Close()

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	kind, ok := fileKind(contentType)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_file_type"})
		return
	}
	if header.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large"})
		return
	}

	id := uuid.New()
	kinesiologistDir := filepath.Join(h.rootDir, kinesiologistID.String())
	if err := os.MkdirAll(kinesiologistDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	storagePath := filepath.Join(kinesiologistDir, id.String()+strings.ToLower(filepath.Ext(header.Filename)))
	out, err := os.Create(storagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	size, copyErr := io.Copy(out, io.LimitReader(file, maxUploadBytes+1))
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil || size > maxUploadBytes {
		_ = os.Remove(storagePath)
		if size > maxUploadBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_too_large"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	attachment := domain.Attachment{
		ID:              id,
		KinesiologistID: kinesiologistID,
		FileName:        filepath.Base(header.Filename),
		ContentType:     contentType,
		SizeBytes:       size,
		StoragePath:     storagePath,
		Kind:            kind,
		Category:        normalizeCategory(c.PostForm("category")),
		Notes:           trimOptionalString(c.PostForm("notes")),
		UploadedByEmail: trimOptionalString(actorEmail(c)),
		UploadedByRole:  trimOptionalString(actorRole(c)),
		CreatedAt:       time.Now().UTC(),
	}

	created, err := h.repo.Create(c.Request.Context(), attachment)
	if err != nil {
		_ = os.Remove(storagePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(created))
}

func (h *Handler) Download(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("attachment_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_attachment_id"})
		return
	}

	attachment, found, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	if _, err := os.Stat(attachment.StoragePath); errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file_not_found"})
		return
	}

	c.Header("Content-Type", attachment.ContentType)
	c.FileAttachment(attachment.StoragePath, attachment.FileName)
}

func fileKind(contentType string) (string, bool) {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image", true
	case strings.HasPrefix(contentType, "video/"):
		return "video", true
	case contentType == "application/pdf":
		return "pdf", true
	default:
		return "", false
	}
}

func normalizeCategory(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "matricula", "titulo", "dni", "seguro", "foto", "video", "documento", "otro":
		return value
	default:
		return "otro"
	}
}

func toResponse(attachment domain.Attachment) attachmentResponse {
	return attachmentResponse{
		ID:              attachment.ID.String(),
		KinesiologistID: attachment.KinesiologistID.String(),
		FileName:        attachment.FileName,
		ContentType:     attachment.ContentType,
		SizeBytes:       attachment.SizeBytes,
		Kind:            attachment.Kind,
		Category:        attachment.Category,
		Notes:           attachment.Notes,
		UploadedByEmail: attachment.UploadedByEmail,
		UploadedByRole:  attachment.UploadedByRole,
		DownloadURL:     "/api/v1/kinesiologist-attachments/" + attachment.ID.String() + "/download",
		CreatedAt:       attachment.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func trimOptionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func actorEmail(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Email
	}
	if middleware.HasRole(c, "recepcionista") {
		return "demo@local"
	}
	return ""
}

func actorRole(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Role
	}
	if middleware.HasRole(c, "recepcionista") {
		return "recepcionista"
	}
	return ""
}
