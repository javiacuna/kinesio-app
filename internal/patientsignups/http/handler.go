package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/javiacuna/kinesio-backend/internal/http/middleware"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/domain"
	"github.com/javiacuna/kinesio-backend/internal/patientsignups/usecase"
)

type Handler struct {
	create   *usecase.CreateSignupRequestUseCase
	approve  *usecase.ApproveSignupRequestUseCase
	reject   *usecase.RejectSignupRequestUseCase
	list     *usecase.ListSignupRequestsUseCase
	myStatus *usecase.GetMySignupStatusUseCase
}

func NewHandler(create *usecase.CreateSignupRequestUseCase, approve *usecase.ApproveSignupRequestUseCase, reject *usecase.RejectSignupRequestUseCase, list *usecase.ListSignupRequestsUseCase, myStatus *usecase.GetMySignupStatusUseCase) *Handler {
	return &Handler{create: create, approve: approve, reject: reject, list: list, myStatus: myStatus}
}

type mySignupStatusResponse struct {
	Status string `json:"status"`
}

type registerAccountRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	DNI       string `json:"dni"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type registerAccountResponse struct {
	Status string `json:"status"`
}

type rejectRequest struct {
	Reason *string `json:"reason"`
}

type signupRequestResponse struct {
	ID               string                 `json:"id"`
	FirebaseUID      string                 `json:"firebase_uid"`
	DNI              string                 `json:"dni"`
	Email            string                 `json:"email"`
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	Status           string                 `json:"status"`
	MatchedPatientID *string                `json:"matched_patient_id,omitempty"`
	ReviewedByEmail  *string                `json:"reviewed_by_email,omitempty"`
	ReviewedAt       *string                `json:"reviewed_at,omitempty"`
	RejectionReason  *string                `json:"rejection_reason,omitempty"`
	CreatedAt        string                 `json:"created_at"`
	PossibleMatch    *possibleMatchResponse `json:"possible_match,omitempty"`
}

// possibleMatchResponse avisa al revisor que el DNI de la solicitud ya
// pertenece a un paciente cargado por la clínica, aunque el email no haya
// coincidido lo suficiente como para auto-aprobarla.
type possibleMatchResponse struct {
	PatientID    string `json:"patient_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	EmailMatches bool   `json:"email_matches"`
}

// MyStatus permite que el propio usuario autenticado (todavía sin rol,
// porque su autorregistro está en revisión) consulte en qué quedó su
// solicitud, en vez de ver únicamente la pantalla genérica de "sin rol
// asignado". No exige ningún rol particular — a propósito, porque el punto
// es que este usuario no tiene ninguno todavía.
func (h *Handler) MyStatus(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	req, found, err := h.myStatus.Execute(c.Request.Context(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !found {
		c.JSON(http.StatusOK, mySignupStatusResponse{Status: "none"})
		return
	}
	c.JSON(http.StatusOK, mySignupStatusResponse{Status: string(req.Status)})
}

func (h *Handler) Register(c *gin.Context) {
	var req registerAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}

	out, validation, err := h.create.Execute(c.Request.Context(), usecase.CreateSignupRequestInput{
		Email:     req.Email,
		Password:  req.Password,
		DNI:       req.DNI,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})

	if err != nil {
		writeSignupError(c, err, validation)
		return
	}

	c.JSON(http.StatusCreated, registerAccountResponse{Status: string(out.Status)})
}

func (h *Handler) List(c *gin.Context) {
	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		status = "pending"
	}
	if status == "all" {
		status = ""
	}

	items, err := h.list.Execute(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	out := make([]signupRequestResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toListItemResponse(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) Approve(c *gin.Context) {
	reviewer := reviewerEmail(c)
	out, err := h.approve.Execute(c.Request.Context(), c.Param("request_id"), reviewer)
	if err != nil {
		writeSignupError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, toResponse(out))
}

func (h *Handler) Reject(c *gin.Context) {
	var req rejectRequest
	_ = c.ShouldBindJSON(&req)

	reviewer := reviewerEmail(c)
	out, err := h.reject.Execute(c.Request.Context(), c.Param("request_id"), reviewer, req.Reason)
	if err != nil {
		writeSignupError(c, err, nil)
		return
	}
	c.JSON(http.StatusOK, toResponse(out))
}

func reviewerEmail(c *gin.Context) string {
	if user, ok := middleware.CurrentUser(c); ok {
		return user.Email
	}
	return "demo@local"
}

func writeSignupError(c *gin.Context, err error, validation map[string]string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
	case errors.Is(err, domain.ErrEmailInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "email_already_registered"})
	case errors.Is(err, domain.ErrAlreadyPending):
		c.JSON(http.StatusConflict, gin.H{"error": "signup_already_pending"})
	case errors.Is(err, domain.ErrDNIAlreadyClaimed):
		c.JSON(http.StatusConflict, gin.H{"error": "dni_already_claimed"})
	case errors.Is(err, domain.ErrRequestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "request_not_found"})
	case errors.Is(err, domain.ErrAlreadyReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": "already_reviewed"})
	case errors.Is(err, domain.ErrPatientConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "patient_conflict"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func toResponse(r domain.SignupRequest) signupRequestResponse {
	var matchedPatientID *string
	if r.MatchedPatientID != nil {
		s := r.MatchedPatientID.String()
		matchedPatientID = &s
	}
	var reviewedAt *string
	if r.ReviewedAt != nil {
		s := r.ReviewedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		reviewedAt = &s
	}

	return signupRequestResponse{
		ID:               r.ID.String(),
		FirebaseUID:      r.FirebaseUID,
		DNI:              r.DNI,
		Email:            r.Email,
		FirstName:        r.FirstName,
		LastName:         r.LastName,
		Status:           string(r.Status),
		MatchedPatientID: matchedPatientID,
		ReviewedByEmail:  r.ReviewedByEmail,
		ReviewedAt:       reviewedAt,
		RejectionReason:  r.RejectionReason,
		CreatedAt:        r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toListItemResponse(item usecase.SignupRequestListItem) signupRequestResponse {
	resp := toResponse(item.SignupRequest)
	if item.PossibleMatch != nil {
		resp.PossibleMatch = &possibleMatchResponse{
			PatientID:    item.PossibleMatch.PatientID.String(),
			FirstName:    item.PossibleMatch.FirstName,
			LastName:     item.PossibleMatch.LastName,
			Email:        item.PossibleMatch.Email,
			EmailMatches: item.PossibleMatch.EmailMatches,
		}
	}
	return resp
}
