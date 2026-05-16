package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/finance/domain"
)

type Store interface {
	ListFinanciers(ctx context.Context, includeInactive bool) ([]domain.Financier, error)
	SaveFinancier(ctx context.Context, item domain.Financier) (domain.Financier, error)
	ListTariffs(ctx context.Context, includeInactive bool) ([]domain.PracticeTariff, error)
	SaveTariff(ctx context.Context, item domain.PracticeTariff) (domain.PracticeTariff, error)
	ListFeeRules(ctx context.Context, includeInactive bool) ([]domain.ProfessionalFeeRule, error)
	SaveFeeRule(ctx context.Context, item domain.ProfessionalFeeRule) (domain.ProfessionalFeeRule, error)
	ListMovements(ctx context.Context, filters domain.MovementFilters) ([]domain.FinancialMovement, error)
	UpdateMovementStatuses(ctx context.Context, id uuid.UUID, collectionStatus *string, professionalPaymentStatus *string, cancellationReason *string) (domain.FinancialMovement, error)
	CompleteAppointment(ctx context.Context, appointmentID, practiceID, financierID uuid.UUID) (domain.FinancialMovement, error)
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

type financierReq struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Active bool   `json:"active"`
}

type financierResp struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type tariffReq struct {
	PracticeID        string  `json:"practice_id"`
	FinancierID       string  `json:"financier_id"`
	BillingValueCents int64   `json:"billing_value_cents"`
	CopayCents        int64   `json:"copay_cents"`
	Currency          string  `json:"currency"`
	ValidFrom         string  `json:"valid_from"`
	ValidTo           *string `json:"valid_to"`
	Active            bool    `json:"active"`
}

type tariffResp struct {
	ID                string  `json:"id"`
	PracticeID        string  `json:"practice_id"`
	FinancierID       string  `json:"financier_id"`
	BillingValueCents int64   `json:"billing_value_cents"`
	CopayCents        int64   `json:"copay_cents"`
	Currency          string  `json:"currency"`
	ValidFrom         string  `json:"valid_from"`
	ValidTo           *string `json:"valid_to,omitempty"`
	Active            bool    `json:"active"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type feeRuleReq struct {
	KinesiologistID string   `json:"kinesiologist_id"`
	PracticeID      string   `json:"practice_id"`
	RuleType        string   `json:"rule_type"`
	FixedValueCents *int64   `json:"fixed_value_cents"`
	Percentage      *float64 `json:"percentage"`
	Active          bool     `json:"active"`
}

type feeRuleResp struct {
	ID              string   `json:"id"`
	KinesiologistID string   `json:"kinesiologist_id"`
	PracticeID      string   `json:"practice_id"`
	RuleType        string   `json:"rule_type"`
	FixedValueCents *int64   `json:"fixed_value_cents,omitempty"`
	Percentage      *float64 `json:"percentage,omitempty"`
	Active          bool     `json:"active"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type completeAppointmentReq struct {
	PracticeID  string `json:"practice_id"`
	FinancierID string `json:"financier_id"`
}

type movementResp struct {
	ID                    string  `json:"id"`
	AppointmentID         string  `json:"appointment_id"`
	PatientID             string  `json:"patient_id"`
	KinesiologistID       string  `json:"kinesiologist_id"`
	PracticeID            string  `json:"practice_id"`
	FinancierID           string  `json:"financier_id"`
	TariffID              string  `json:"tariff_id"`
	FeeRuleID             *string `json:"fee_rule_id,omitempty"`
	BillingValueCents     int64   `json:"billing_value_cents"`
	CopayCents            int64   `json:"copay_cents"`
	PayerValueCents       int64   `json:"payer_value_cents"`
	ProfessionalFeeCents  int64   `json:"professional_fee_cents"`
	CenterAmountCents     int64   `json:"center_amount_cents"`
	Status                string  `json:"status"`
	CollectionStatus      string  `json:"collection_status"`
	ProfessionalPayStatus string  `json:"professional_payment_status"`
	CollectedAt           *string `json:"collected_at,omitempty"`
	ProfessionalPaidAt    *string `json:"professional_paid_at,omitempty"`
	CancellationReason    *string `json:"cancellation_reason,omitempty"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

type movementStatusReq struct {
	CollectionStatus          *string `json:"collection_status"`
	ProfessionalPaymentStatus *string `json:"professional_payment_status"`
	CancellationReason        *string `json:"cancellation_reason"`
}

func (h *Handler) ListFinanciers(c *gin.Context) {
	items, err := h.store.ListFinanciers(c.Request.Context(), includeInactive(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]financierResp, 0, len(items))
	for _, item := range items {
		out = append(out, toFinancierResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateFinancier(c *gin.Context) {
	var req financierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildFinancier("", req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveFinancier(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, toFinancierResp(out))
}

func (h *Handler) UpdateFinancier(c *gin.Context) {
	var req financierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildFinancier(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveFinancier(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, toFinancierResp(out))
}

func (h *Handler) ListTariffs(c *gin.Context) {
	items, err := h.store.ListTariffs(c.Request.Context(), includeInactive(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]tariffResp, 0, len(items))
	for _, item := range items {
		out = append(out, toTariffResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateTariff(c *gin.Context) {
	var req tariffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildTariff("", req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveTariff(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, toTariffResp(out))
}

func (h *Handler) UpdateTariff(c *gin.Context) {
	var req tariffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildTariff(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveTariff(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, toTariffResp(out))
}

func (h *Handler) ListFeeRules(c *gin.Context) {
	items, err := h.store.ListFeeRules(c.Request.Context(), includeInactive(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]feeRuleResp, 0, len(items))
	for _, item := range items {
		out = append(out, toFeeRuleResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) CreateFeeRule(c *gin.Context) {
	var req feeRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildFeeRule("", req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveFeeRule(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, toFeeRuleResp(out))
}

func (h *Handler) UpdateFeeRule(c *gin.Context) {
	var req feeRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	item, validation, err := buildFeeRule(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": validation})
		return
	}
	out, err := h.store.SaveFeeRule(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, toFeeRuleResp(out))
}

func (h *Handler) ListMovements(c *gin.Context) {
	from, fromOK := parseOptionalRFC3339(c.Query("from"))
	to, toOK := parseOptionalRFC3339(c.Query("to"))
	if !fromOK || !toOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	items, err := h.store.ListMovements(
		c.Request.Context(),
		domain.MovementFilters{
			From:                      from,
			To:                        to,
			Status:                    strings.TrimSpace(c.Query("status")),
			KinesiologistID:           strings.TrimSpace(c.Query("kinesiologist_id")),
			PracticeID:                strings.TrimSpace(c.Query("practice_id")),
			FinancierID:               strings.TrimSpace(c.Query("financier_id")),
			CollectionStatus:          strings.TrimSpace(c.Query("collection_status")),
			ProfessionalPaymentStatus: strings.TrimSpace(c.Query("professional_payment_status")),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]movementResp, 0, len(items))
	for _, item := range items {
		out = append(out, toMovementResp(item))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) UpdateMovementStatuses(c *gin.Context) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param("movement_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"id": "UUID inválido"}})
		return
	}
	var req movementStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	if req.CollectionStatus == nil && req.ProfessionalPaymentStatus == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"status": "Debe enviar al menos un estado"}})
		return
	}
	if req.CollectionStatus != nil && !validCollectionStatus(*req.CollectionStatus) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"collection_status": "Valor inválido"}})
		return
	}
	if req.ProfessionalPaymentStatus != nil && !validProfessionalPaymentStatus(*req.ProfessionalPaymentStatus) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"professional_payment_status": "Valor inválido"}})
		return
	}
	if isCancelling(req.CollectionStatus, req.ProfessionalPaymentStatus) && strings.TrimSpace(optionalString(req.CancellationReason)) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"cancellation_reason": "Completá el motivo de anulación"}})
		return
	}
	reason := trimOptionalString(req.CancellationReason)
	out, err := h.store.UpdateMovementStatuses(c.Request.Context(), id, req.CollectionStatus, req.ProfessionalPaymentStatus, reason)
	if err != nil {
		writeFinanceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMovementResp(out))
}

func (h *Handler) CompleteAppointment(c *gin.Context) {
	var req completeAppointmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json"})
		return
	}
	appointmentID, err := uuid.Parse(strings.TrimSpace(c.Param("appointment_id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"appointment_id": "UUID inválido"}})
		return
	}
	practiceID, err := uuid.Parse(strings.TrimSpace(req.PracticeID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"practice_id": "UUID inválido"}})
		return
	}
	financierID, err := uuid.Parse(strings.TrimSpace(req.FinancierID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": gin.H{"financier_id": "UUID inválido"}})
		return
	}
	out, err := h.store.CompleteAppointment(c.Request.Context(), appointmentID, practiceID, financierID)
	if err != nil {
		writeFinanceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toMovementResp(out))
}

func buildFinancier(id string, req financierReq) (domain.Financier, map[string]string, error) {
	errs := map[string]string{}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		errs["name"] = "Campo obligatorio"
	}
	kind := strings.TrimSpace(req.Kind)
	if kind == "" {
		kind = "particular"
	}
	parsedID := uuid.New()
	if strings.TrimSpace(id) != "" {
		var err error
		parsedID, err = uuid.Parse(strings.TrimSpace(id))
		if err != nil {
			errs["id"] = "UUID inválido"
		}
	}
	if len(errs) > 0 {
		return domain.Financier{}, errs, domain.ErrValidation
	}
	return domain.Financier{ID: parsedID, Name: name, Kind: kind, Active: req.Active}, nil, nil
}

func buildTariff(id string, req tariffReq) (domain.PracticeTariff, map[string]string, error) {
	errs := map[string]string{}
	practiceID, err := uuid.Parse(strings.TrimSpace(req.PracticeID))
	if err != nil {
		errs["practice_id"] = "UUID inválido"
	}
	financierID, err := uuid.Parse(strings.TrimSpace(req.FinancierID))
	if err != nil {
		errs["financier_id"] = "UUID inválido"
	}
	validFrom, err := time.Parse("2006-01-02", strings.TrimSpace(req.ValidFrom))
	if err != nil {
		errs["valid_from"] = "Fecha inválida"
	}
	var validTo *time.Time
	if req.ValidTo != nil && strings.TrimSpace(*req.ValidTo) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ValidTo))
		if err != nil {
			errs["valid_to"] = "Fecha inválida"
		} else {
			validTo = &parsed
		}
	}
	if req.BillingValueCents < 0 {
		errs["billing_value_cents"] = "Debe ser mayor o igual a 0"
	}
	if req.CopayCents < 0 {
		errs["copay_cents"] = "Debe ser mayor o igual a 0"
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "ARS"
	}
	parsedID := uuid.New()
	if strings.TrimSpace(id) != "" {
		parsedID, err = uuid.Parse(strings.TrimSpace(id))
		if err != nil {
			errs["id"] = "UUID inválido"
		}
	}
	if len(errs) > 0 {
		return domain.PracticeTariff{}, errs, domain.ErrValidation
	}
	return domain.PracticeTariff{
		ID:                parsedID,
		PracticeID:        practiceID,
		FinancierID:       financierID,
		BillingValueCents: req.BillingValueCents,
		CopayCents:        req.CopayCents,
		Currency:          currency,
		ValidFrom:         validFrom,
		ValidTo:           validTo,
		Active:            req.Active,
	}, nil, nil
}

func buildFeeRule(id string, req feeRuleReq) (domain.ProfessionalFeeRule, map[string]string, error) {
	errs := map[string]string{}
	kinesiologistID, err := uuid.Parse(strings.TrimSpace(req.KinesiologistID))
	if err != nil {
		errs["kinesiologist_id"] = "UUID inválido"
	}
	practiceID, err := uuid.Parse(strings.TrimSpace(req.PracticeID))
	if err != nil {
		errs["practice_id"] = "UUID inválido"
	}
	ruleType := strings.TrimSpace(req.RuleType)
	if ruleType != "fixed" && ruleType != "percentage" {
		errs["rule_type"] = "Valor inválido"
	}
	if ruleType == "fixed" && (req.FixedValueCents == nil || *req.FixedValueCents < 0) {
		errs["fixed_value_cents"] = "Completá un monto fijo válido"
	}
	if ruleType == "percentage" && (req.Percentage == nil || *req.Percentage < 0 || *req.Percentage > 100) {
		errs["percentage"] = "Completá un porcentaje entre 0 y 100"
	}
	fixed := req.FixedValueCents
	percentage := req.Percentage
	if ruleType == "fixed" {
		percentage = nil
	}
	if ruleType == "percentage" {
		fixed = nil
	}
	parsedID := uuid.New()
	if strings.TrimSpace(id) != "" {
		parsedID, err = uuid.Parse(strings.TrimSpace(id))
		if err != nil {
			errs["id"] = "UUID inválido"
		}
	}
	if len(errs) > 0 {
		return domain.ProfessionalFeeRule{}, errs, domain.ErrValidation
	}
	return domain.ProfessionalFeeRule{
		ID:              parsedID,
		KinesiologistID: kinesiologistID,
		PracticeID:      practiceID,
		RuleType:        ruleType,
		FixedValueCents: fixed,
		Percentage:      percentage,
		Active:          req.Active,
	}, nil, nil
}

func includeInactive(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.Query("active")), "false")
}

func parseOptionalRFC3339(value string) (*time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

func writeFinanceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	case errors.Is(err, domain.ErrTariffNotFound):
		c.JSON(http.StatusConflict, gin.H{"error": "tariff_not_found"})
	case errors.Is(err, domain.ErrAlreadyGenerated):
		c.JSON(http.StatusConflict, gin.H{"error": "financial_movement_already_generated"})
	case errors.Is(err, domain.ErrInvalidStatus):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid_appointment_status"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func validCollectionStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "pending", "collected", "cancelled":
		return true
	default:
		return false
	}
}

func validProfessionalPaymentStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "pending", "paid", "cancelled":
		return true
	default:
		return false
	}
}

func isCancelling(collectionStatus *string, professionalPaymentStatus *string) bool {
	return (collectionStatus != nil && strings.TrimSpace(*collectionStatus) == "cancelled") ||
		(professionalPaymentStatus != nil && strings.TrimSpace(*professionalPaymentStatus) == "cancelled")
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toFinancierResp(item domain.Financier) financierResp {
	return financierResp{
		ID:        item.ID.String(),
		Name:      item.Name,
		Kind:      item.Kind,
		Active:    item.Active,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.Format(time.RFC3339),
	}
}

func toTariffResp(item domain.PracticeTariff) tariffResp {
	var validTo *string
	if item.ValidTo != nil {
		value := item.ValidTo.Format("2006-01-02")
		validTo = &value
	}
	return tariffResp{
		ID:                item.ID.String(),
		PracticeID:        item.PracticeID.String(),
		FinancierID:       item.FinancierID.String(),
		BillingValueCents: item.BillingValueCents,
		CopayCents:        item.CopayCents,
		Currency:          item.Currency,
		ValidFrom:         item.ValidFrom.Format("2006-01-02"),
		ValidTo:           validTo,
		Active:            item.Active,
		CreatedAt:         item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         item.UpdatedAt.Format(time.RFC3339),
	}
}

func toFeeRuleResp(item domain.ProfessionalFeeRule) feeRuleResp {
	return feeRuleResp{
		ID:              item.ID.String(),
		KinesiologistID: item.KinesiologistID.String(),
		PracticeID:      item.PracticeID.String(),
		RuleType:        item.RuleType,
		FixedValueCents: item.FixedValueCents,
		Percentage:      item.Percentage,
		Active:          item.Active,
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.Format(time.RFC3339),
	}
}

func toMovementResp(item domain.FinancialMovement) movementResp {
	var feeRuleID *string
	if item.FeeRuleID != nil {
		value := item.FeeRuleID.String()
		feeRuleID = &value
	}
	var collectedAt *string
	if item.CollectedAt != nil {
		value := item.CollectedAt.UTC().Format(time.RFC3339)
		collectedAt = &value
	}
	var professionalPaidAt *string
	if item.ProfessionalPaidAt != nil {
		value := item.ProfessionalPaidAt.UTC().Format(time.RFC3339)
		professionalPaidAt = &value
	}
	return movementResp{
		ID:                    item.ID.String(),
		AppointmentID:         item.AppointmentID.String(),
		PatientID:             item.PatientID.String(),
		KinesiologistID:       item.KinesiologistID.String(),
		PracticeID:            item.PracticeID.String(),
		FinancierID:           item.FinancierID.String(),
		TariffID:              item.TariffID.String(),
		FeeRuleID:             feeRuleID,
		BillingValueCents:     item.BillingValueCents,
		CopayCents:            item.CopayCents,
		PayerValueCents:       item.PayerValueCents,
		ProfessionalFeeCents:  item.ProfessionalFeeCents,
		CenterAmountCents:     item.CenterAmountCents,
		Status:                item.Status,
		CollectionStatus:      valueOrDefault(item.CollectionStatus, "pending"),
		ProfessionalPayStatus: valueOrDefault(item.ProfessionalPayStatus, "pending"),
		CollectedAt:           collectedAt,
		ProfessionalPaidAt:    professionalPaidAt,
		CancellationReason:    item.CancellationReason,
		CreatedAt:             item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             item.UpdatedAt.Format(time.RFC3339),
	}
}

func valueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
