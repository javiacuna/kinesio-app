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
	"gorm.io/gorm"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patientattachments/domain"
	patientDomain "github.com/javiacuna/kinesio-backend/internal/patients/domain"
)

const maxUploadBytes = 50 << 20

type repository interface {
	Create(ctx context.Context, attachment domain.Attachment) (domain.Attachment, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.Attachment, error)
	ListVisibleByPatient(ctx context.Context, patientID uuid.UUID) ([]domain.Attachment, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Attachment, bool, error)
	Update(ctx context.Context, attachment domain.Attachment) (domain.Attachment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type patientSearcher interface {
	Search(ctx context.Context, query string, limit int, includeInactive bool) ([]patientDomain.Patient, error)
}

type Handler struct {
	repo     repository
	rootDir  string
	patients patientSearcher
}

func NewHandler(repo repository, rootDir string, lookups ...any) *Handler {
	rootDir = strings.TrimSpace(rootDir)
	if rootDir == "" {
		rootDir = "uploads/patient-attachments"
	}

	var patientRepo patientSearcher
	for _, lookup := range lookups {
		if repo, ok := lookup.(patientSearcher); ok {
			patientRepo = repo
		}
	}

	return &Handler{repo: repo, rootDir: filepath.Clean(rootDir), patients: patientRepo}
}

type attachmentResponse struct {
	ID              string  `json:"id"`
	PatientID       string  `json:"patient_id"`
	FileName        string  `json:"file_name"`
	ContentType     string  `json:"content_type"`
	SizeBytes       int64   `json:"size_bytes"`
	Kind            string  `json:"kind"`
	Category        string  `json:"category"`
	PatientVisible  bool    `json:"patient_visible"`
	Notes           *string `json:"notes,omitempty"`
	UploadedByEmail *string `json:"uploaded_by_email,omitempty"`
	UploadedByRole  *string `json:"uploaded_by_role,omitempty"`
	UpdatedByEmail  *string `json:"updated_by_email,omitempty"`
	UpdatedByRole   *string `json:"updated_by_role,omitempty"`
	UpdatedAt       *string `json:"updated_at,omitempty"`
	DownloadURL     string  `json:"download_url"`
	CreatedAt       string  `json:"created_at"`
}

type updateAttachmentRequest struct {
	FileName       string  `json:"file_name"`
	Category       string  `json:"category"`
	PatientVisible bool    `json:"patient_visible"`
	Notes          *string `json:"notes"`
}

func (h *Handler) ListByPatient(c *gin.Context) {
	patientIDParam := strings.TrimSpace(c.Param("patient_id"))
	visibleOnly := false

	if strings.EqualFold(patientIDParam, "me") || h.isCurrentPatient(c) {
		resolved, ok := h.patientIDForCurrentPatient(c, patientIDParam)
		if !ok {
			return
		}
		patientIDParam = resolved
		visibleOnly = true
	} else if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	patientID, err := uuid.Parse(patientIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
		return
	}

	var items []domain.Attachment
	if visibleOnly {
		items, err = h.repo.ListVisibleByPatient(c.Request.Context(), patientID)
	} else {
		items, err = h.repo.ListByPatient(c.Request.Context(), patientID)
	}
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

	patientID, err := uuid.Parse(strings.TrimSpace(c.Param("patient_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_patient_id"})
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
	patientDir := filepath.Join(h.rootDir, patientID.String())
	if err := os.MkdirAll(patientDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	storagePath := filepath.Join(patientDir, id.String()+strings.ToLower(filepath.Ext(header.Filename)))
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
		PatientID:       patientID,
		FileName:        filepath.Base(header.Filename),
		ContentType:     contentType,
		SizeBytes:       size,
		StoragePath:     storagePath,
		Kind:            kind,
		Category:        normalizeCategory(c.PostForm("category")),
		PatientVisible:  parseBool(c.PostForm("patient_visible")),
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

func (h *Handler) Update(c *gin.Context) {
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := uuid.Parse(strings.TrimSpace(c.Param("attachment_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_attachment_id"})
		return
	}

	var req updateAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	current, found, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" {
		fileName = current.FileName
	}
	now := time.Now().UTC()
	current.FileName = filepath.Base(fileName)
	current.Category = normalizeCategory(req.Category)
	current.PatientVisible = req.PatientVisible
	current.Notes = trimOptionalStringPtr(req.Notes)
	current.UpdatedByEmail = trimOptionalString(actorEmail(c))
	current.UpdatedByRole = trimOptionalString(actorRole(c))
	current.UpdatedAt = &now

	updated, err := h.repo.Update(c.Request.Context(), current)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

func (h *Handler) Delete(c *gin.Context) {
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

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	_ = os.Remove(attachment.StoragePath)

	c.Status(http.StatusNoContent)
}

func (h *Handler) Download(c *gin.Context) {
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
	if !middleware.HasRole(c, "recepcionista", "kinesiologo") && !h.canCurrentPatientAccessAttachment(c, attachment) {
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

func toResponse(attachment domain.Attachment) attachmentResponse {
	var updatedAt *string
	if attachment.UpdatedAt != nil {
		value := attachment.UpdatedAt.UTC().Format(time.RFC3339)
		updatedAt = &value
	}

	return attachmentResponse{
		ID:              attachment.ID.String(),
		PatientID:       attachment.PatientID.String(),
		FileName:        attachment.FileName,
		ContentType:     attachment.ContentType,
		SizeBytes:       attachment.SizeBytes,
		Kind:            attachment.Kind,
		Category:        attachment.Category,
		PatientVisible:  attachment.PatientVisible,
		Notes:           attachment.Notes,
		UploadedByEmail: attachment.UploadedByEmail,
		UploadedByRole:  attachment.UploadedByRole,
		UpdatedByEmail:  attachment.UpdatedByEmail,
		UpdatedByRole:   attachment.UpdatedByRole,
		UpdatedAt:       updatedAt,
		DownloadURL:     "/api/v1/patient-attachments/" + attachment.ID.String() + "/download",
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

func trimOptionalStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	return trimOptionalString(*value)
}

func normalizeCategory(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "radiografia", "resonancia", "ecografia", "laboratorio", "foto", "video", "documento", "otro":
		return value
	default:
		return "otro"
	}
}

func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes" || value == "on"
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

func (h *Handler) patientIDForCurrentPatient(c *gin.Context, requestedPatientID string) (string, bool) {
	user, ok := middleware.CurrentUser(c)
	if !ok || !strings.EqualFold(user.Role, "paciente") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	if h.patients == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	email := strings.TrimSpace(user.Email)
	if email == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "patient_profile_not_found"})
		return "", false
	}

	patients, err := h.patients.Search(c.Request.Context(), email, 10, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return "", false
	}

	for _, patient := range patients {
		if strings.EqualFold(patient.Email, email) && patient.Active {
			ownID := patient.ID.String()
			requested := strings.TrimSpace(requestedPatientID)
			if requested != "" && !strings.EqualFold(requested, "me") && requested != ownID {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return "", false
			}
			return ownID, true
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "patient_profile_not_found"})
	return "", false
}

func (h *Handler) canCurrentPatientAccessAttachment(c *gin.Context, attachment domain.Attachment) bool {
	patientID, ok := h.patientIDForCurrentPatient(c, "")
	if !ok {
		return false
	}
	if attachment.PatientID.String() != patientID || !attachment.PatientVisible {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	return true
}

func (h *Handler) isCurrentPatient(c *gin.Context) bool {
	user, ok := middleware.CurrentUser(c)
	return ok && strings.EqualFold(user.Role, "paciente")
}
