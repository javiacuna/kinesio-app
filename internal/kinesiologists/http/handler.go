package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/domain"
	"github.com/javiacuna/kinesio-backend/internal/kinesiologists/usecase"
)

type Handler struct {
	list            *usecase.ListKinesiologistsUseCase
	save            *usecase.SaveKinesiologistUseCase
	listSpecialties *usecase.ListSpecialtiesUseCase
	saveSpecialty   *usecase.SaveSpecialtyUseCase
	listPractices   *usecase.ListPracticesUseCase
	savePractice    *usecase.SavePracticeUseCase
}

func NewHandler(
	list *usecase.ListKinesiologistsUseCase,
	save *usecase.SaveKinesiologistUseCase,
	listSpecialties *usecase.ListSpecialtiesUseCase,
	saveSpecialty *usecase.SaveSpecialtyUseCase,
	listPractices *usecase.ListPracticesUseCase,
	savePractice *usecase.SavePracticeUseCase,
) *Handler {
	return &Handler{
		list:            list,
		save:            save,
		listSpecialties: listSpecialties,
		saveSpecialty:   saveSpecialty,
		listPractices:   listPractices,
		savePractice:    savePractice,
	}
}

type saveReq struct {
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	LicenseNumber *string  `json:"license_number,omitempty"`
	WorkStartTime string   `json:"work_start_time"`
	WorkEndTime   string   `json:"work_end_time"`
	WorkDays      []int    `json:"work_days"`
	PracticeIDs   []string `json:"practice_ids"`
	Active        bool     `json:"active"`
}

type resp struct {
	ID            string         `json:"id"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Email         string         `json:"email"`
	LicenseNumber *string        `json:"license_number,omitempty"`
	WorkStartTime string         `json:"work_start_time"`
	WorkEndTime   string         `json:"work_end_time"`
	WorkDays      []int          `json:"work_days"`
	PracticeIDs   []string       `json:"practice_ids"`
	Practices     []practiceResp `json:"practices"`
	Active        bool           `json:"active"`
}

type specialtyReq struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type specialtyResp struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type practiceReq struct {
	SpecialtyID string  `json:"specialty_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Active      bool    `json:"active"`
}

type practiceResp struct {
	ID          string  `json:"id"`
	SpecialtyID string  `json:"specialty_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Active      bool    `json:"active"`
}

func (h *Handler) List(c *gin.Context) {
	onlyActive := true
	if v := strings.TrimSpace(c.Query("active")); v != "" {
		// active=true|false; si te pasan false, devolvemos todos
		onlyActive = strings.EqualFold(v, "true")
	}

	if user, ok := middleware.CurrentUser(c); ok && strings.EqualFold(user.Role, "kinesiologo") {
		item, found, err := h.list.FindByEmail(c.Request.Context(), user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		if !found || !item.Active {
			c.JSON(http.StatusOK, []resp{})
			return
		}
		c.JSON(http.StatusOK, []resp{toResp(item)})
		return
	}

	items, err := h.list.Execute(c.Request.Context(), onlyActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]resp, 0, len(items))
	for _, k := range items {
		out = append(out, toResp(k))
	}

	c.JSON(http.StatusOK, out)
}

func (h *Handler) Create(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req saveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Create(c.Request.Context(), usecase.SaveKinesiologistInput{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		LicenseNumber: req.LicenseNumber,
		WorkStartTime: req.WorkStartTime,
		WorkEndTime:   req.WorkEndTime,
		WorkDays:      req.WorkDays,
		PracticeIDs:   req.PracticeIDs,
		Active:        req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}

	c.JSON(http.StatusCreated, toResp(out))
}

func (h *Handler) Update(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req saveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.save.Update(c.Request.Context(), usecase.SaveKinesiologistInput{
		ID:            c.Param("id"),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		LicenseNumber: req.LicenseNumber,
		WorkStartTime: req.WorkStartTime,
		WorkEndTime:   req.WorkEndTime,
		WorkDays:      req.WorkDays,
		PracticeIDs:   req.PracticeIDs,
		Active:        req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}

	c.JSON(http.StatusOK, toResp(out))
}

func writeSaveError(c *gin.Context, validation map[string]string, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
	case errors.Is(err, domain.ErrDuplicateEmail):
		c.JSON(http.StatusConflict, gin.H{"error": "email_duplicado"})
	case errors.Is(err, domain.ErrDuplicateName):
		c.JSON(http.StatusConflict, gin.H{"error": "nombre_duplicado"})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func (h *Handler) ListSpecialties(c *gin.Context) {
	includeInactive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")
	items, err := h.listSpecialties.Execute(c.Request.Context(), includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]specialtyResp, 0, len(items))
	for _, item := range items {
		out = append(out, toSpecialtyResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateSpecialty(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req specialtyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	out, validation, err := h.saveSpecialty.Create(c.Request.Context(), usecase.SaveSpecialtyInput{
		Name:   req.Name,
		Active: req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusCreated, toSpecialtyResp(out))
}

func (h *Handler) UpdateSpecialty(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req specialtyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	out, validation, err := h.saveSpecialty.Update(c.Request.Context(), usecase.SaveSpecialtyInput{
		ID:     c.Param("id"),
		Name:   req.Name,
		Active: req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusOK, toSpecialtyResp(out))
}

func (h *Handler) ListPractices(c *gin.Context) {
	includeInactive := strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")
	items, err := h.listPractices.Execute(c.Request.Context(), includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]practiceResp, 0, len(items))
	for _, item := range items {
		out = append(out, toPracticeResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreatePractice(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req practiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	out, validation, err := h.savePractice.Create(c.Request.Context(), usecase.SavePracticeInput{
		SpecialtyID: req.SpecialtyID,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusCreated, toPracticeResp(out))
}

func (h *Handler) UpdatePractice(c *gin.Context) {
	if !middleware.HasRole(c, "admin") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req practiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	out, validation, err := h.savePractice.Update(c.Request.Context(), usecase.SavePracticeInput{
		ID:          c.Param("id"),
		SpecialtyID: req.SpecialtyID,
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	})
	if err != nil {
		writeSaveError(c, validation, err)
		return
	}
	c.JSON(http.StatusOK, toPracticeResp(out))
}

func toResp(k domain.Kinesiologist) resp {
	practiceIDs := make([]string, 0, len(k.PracticeIDs))
	for _, id := range k.PracticeIDs {
		practiceIDs = append(practiceIDs, id.String())
	}
	practices := make([]practiceResp, 0, len(k.Practices))
	for _, practice := range k.Practices {
		practices = append(practices, toPracticeResp(practice))
	}
	return resp{
		ID:            k.ID.String(),
		FirstName:     k.FirstName,
		LastName:      k.LastName,
		Email:         k.Email,
		LicenseNumber: k.LicenseNumber,
		WorkStartTime: k.WorkStartTime,
		WorkEndTime:   k.WorkEndTime,
		WorkDays:      k.WorkDays,
		PracticeIDs:   practiceIDs,
		Practices:     practices,
		Active:        k.Active,
	}
}

func toSpecialtyResp(s domain.Specialty) specialtyResp {
	return specialtyResp{
		ID:     s.ID.String(),
		Name:   s.Name,
		Active: s.Active,
	}
}

func toPracticeResp(p domain.Practice) practiceResp {
	return practiceResp{
		ID:          p.ID.String(),
		SpecialtyID: p.SpecialtyID.String(),
		Name:        p.Name,
		Description: p.Description,
		Active:      p.Active,
	}
}
