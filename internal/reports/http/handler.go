package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

type summaryResponse struct {
	Period                      periodResponse             `json:"period"`
	Totals                      totalsResponse             `json:"totals"`
	AppointmentsByDay           []appointmentDayResponse   `json:"appointments_by_day"`
	RevenueByDay                []revenueDayResponse       `json:"revenue_by_day"`
	AppointmentsByStatus        []labelValueResponse       `json:"appointments_by_status"`
	AppointmentsByKinesiologist []labelValueResponse       `json:"appointments_by_kinesiologist"`
	LowStockMaterials           []lowStockMaterialResponse `json:"low_stock_materials"`
	PendingLoansByMaterial      []labelValueResponse       `json:"pending_loans_by_material"`
}

type periodResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type totalsResponse struct {
	AppointmentsTotal      int   `json:"appointments_total"`
	ScheduledAppointments  int   `json:"scheduled_appointments"`
	CancelledAppointments  int   `json:"cancelled_appointments"`
	ActivePatients         int   `json:"active_patients"`
	LowStockMaterials      int   `json:"low_stock_materials"`
	PendingMaterialLoans   int   `json:"pending_material_loans"`
	BillingValueCents      int64 `json:"billing_value_cents"`
	CollectedCents         int64 `json:"collected_cents"`
	PendingCollectionCents int64 `json:"pending_collection_cents"`
	ProfessionalFeeCents   int64 `json:"professional_fee_cents"`
	CenterAmountCents      int64 `json:"center_amount_cents"`
}

type appointmentDayResponse struct {
	Date      string `json:"date"`
	Scheduled int    `json:"scheduled"`
	Cancelled int    `json:"cancelled"`
	Total     int    `json:"total"`
}

type revenueDayResponse struct {
	Date              string `json:"date"`
	BillingValueCents int64  `json:"billing_value_cents"`
	CollectedCents    int64  `json:"collected_cents"`
}

type labelValueResponse struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type lowStockMaterialResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AvailableQty int    `json:"available_qty"`
	TotalQty     int    `json:"total_qty"`
}

func (h *Handler) Summary(c *gin.Context) {
	from, to, ok := parsePeriod(c)
	if !ok {
		return
	}

	var totals totalsResponse

	var appointmentTotals struct {
		AppointmentsTotal     int `json:"appointments_total"`
		ScheduledAppointments int `json:"scheduled_appointments"`
		CancelledAppointments int `json:"cancelled_appointments"`
	}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			COUNT(*)::int AS appointments_total,
			COUNT(*) FILTER (WHERE status = 'scheduled')::int AS scheduled_appointments,
			COUNT(*) FILTER (WHERE status = 'cancelled')::int AS cancelled_appointments
		FROM appointments
		WHERE start_at >= ? AND start_at < ?
	`, from, to).Scan(&appointmentTotals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	totals.AppointmentsTotal = appointmentTotals.AppointmentsTotal
	totals.ScheduledAppointments = appointmentTotals.ScheduledAppointments
	totals.CancelledAppointments = appointmentTotals.CancelledAppointments

	var resourceTotals struct {
		ActivePatients       int `json:"active_patients"`
		LowStockMaterials    int `json:"low_stock_materials"`
		PendingMaterialLoans int `json:"pending_material_loans"`
	}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			(SELECT COUNT(*)::int FROM patients WHERE active = true) AS active_patients,
			(SELECT COUNT(*)::int FROM materials WHERE total_qty > 0 AND available_qty <= 1) AS low_stock_materials,
			(SELECT COUNT(*)::int FROM material_loans WHERE returned_at IS NULL) AS pending_material_loans
	`).Scan(&resourceTotals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	totals.ActivePatients = resourceTotals.ActivePatients
	totals.LowStockMaterials = resourceTotals.LowStockMaterials
	totals.PendingMaterialLoans = resourceTotals.PendingMaterialLoans

	var financialTotals struct {
		BillingValueCents      int64 `json:"billing_value_cents"`
		ProfessionalFeeCents   int64 `json:"professional_fee_cents"`
		CenterAmountCents      int64 `json:"center_amount_cents"`
		CollectedCents         int64 `json:"collected_cents"`
		PendingCollectionCents int64 `json:"pending_collection_cents"`
	}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			COALESCE(SUM(billing_value_cents), 0)::bigint AS billing_value_cents,
			COALESCE(SUM(professional_fee_cents), 0)::bigint AS professional_fee_cents,
			COALESCE(SUM(center_amount_cents), 0)::bigint AS center_amount_cents,
			COALESCE(SUM(CASE WHEN collection_status = 'collected' THEN billing_value_cents ELSE 0 END), 0)::bigint AS collected_cents,
			COALESCE(SUM(CASE WHEN collection_status <> 'collected' THEN billing_value_cents ELSE 0 END), 0)::bigint AS pending_collection_cents
		FROM financial_movements
		WHERE created_at >= ? AND created_at < ?
	`, from, to).Scan(&financialTotals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	totals.BillingValueCents = financialTotals.BillingValueCents
	totals.ProfessionalFeeCents = financialTotals.ProfessionalFeeCents
	totals.CenterAmountCents = financialTotals.CenterAmountCents
	totals.CollectedCents = financialTotals.CollectedCents
	totals.PendingCollectionCents = financialTotals.PendingCollectionCents

	appointmentsByDay := []appointmentDayResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			to_char(date_trunc('day', start_at AT TIME ZONE 'America/Argentina/Cordoba'), 'YYYY-MM-DD') AS date,
			COUNT(*) FILTER (WHERE status = 'scheduled')::int AS scheduled,
			COUNT(*) FILTER (WHERE status = 'cancelled')::int AS cancelled,
			COUNT(*)::int AS total
		FROM appointments
		WHERE start_at >= ? AND start_at < ?
		GROUP BY 1
		ORDER BY 1
	`, from, to).Scan(&appointmentsByDay).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	revenueByDay := []revenueDayResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			to_char(date_trunc('day', created_at AT TIME ZONE 'America/Argentina/Cordoba'), 'YYYY-MM-DD') AS date,
			COALESCE(SUM(billing_value_cents), 0)::bigint AS billing_value_cents,
			COALESCE(SUM(CASE WHEN collection_status = 'collected' THEN billing_value_cents ELSE 0 END), 0)::bigint AS collected_cents
		FROM financial_movements
		WHERE created_at >= ? AND created_at < ?
		GROUP BY 1
		ORDER BY 1
	`, from, to).Scan(&revenueByDay).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	appointmentsByStatus := []labelValueResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT status AS label, COUNT(*)::int AS value
		FROM appointments
		WHERE start_at >= ? AND start_at < ?
		GROUP BY status
		ORDER BY value DESC
	`, from, to).Scan(&appointmentsByStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	appointmentsByKinesiologist := []labelValueResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT
			TRIM(k.last_name || ', ' || k.first_name) AS label,
			COUNT(a.id)::int AS value
		FROM appointments a
		JOIN kinesiologists k ON k.id = a.kinesiologist_id
		WHERE a.start_at >= ? AND a.start_at < ?
		GROUP BY k.id, k.last_name, k.first_name
		ORDER BY value DESC, label ASC
		LIMIT 10
	`, from, to).Scan(&appointmentsByKinesiologist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	lowStockMaterials := []lowStockMaterialResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT id, name, available_qty, total_qty
		FROM materials
		WHERE total_qty > 0 AND available_qty <= 1
		ORDER BY available_qty ASC, name ASC
		LIMIT 10
	`).Scan(&lowStockMaterials).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	pendingLoansByMaterial := []labelValueResponse{}
	if err := h.db.WithContext(c.Request.Context()).Raw(`
		SELECT m.name AS label, COALESCE(SUM(ml.qty), 0)::int AS value
		FROM material_loans ml
		JOIN materials m ON m.id = ml.material_id
		WHERE ml.returned_at IS NULL
		GROUP BY m.id, m.name
		ORDER BY value DESC, label ASC
		LIMIT 10
	`).Scan(&pendingLoansByMaterial).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, summaryResponse{
		Period: periodResponse{
			From: from.Format("2006-01-02"),
			To:   to.AddDate(0, 0, -1).Format("2006-01-02"),
		},
		Totals:                      totals,
		AppointmentsByDay:           appointmentsByDay,
		RevenueByDay:                revenueByDay,
		AppointmentsByStatus:        appointmentsByStatus,
		AppointmentsByKinesiologist: appointmentsByKinesiologist,
		LowStockMaterials:           lowStockMaterials,
		PendingLoansByMaterial:      pendingLoansByMaterial,
	})
}

func parsePeriod(c *gin.Context) (time.Time, time.Time, bool) {
	loc, err := time.LoadLocation("America/Argentina/Cordoba")
	if err != nil {
		loc = time.FixedZone("ART", -3*60*60)
	}

	now := time.Now().In(loc)
	defaultFrom := now.AddDate(0, 0, -29)
	fromDate := strings.TrimSpace(c.Query("from"))
	toDate := strings.TrimSpace(c.Query("to"))
	if fromDate == "" {
		fromDate = defaultFrom.Format("2006-01-02")
	}
	if toDate == "" {
		toDate = now.Format("2006-01-02")
	}

	from, err := time.ParseInLocation("2006-01-02", fromDate, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_from"})
		return time.Time{}, time.Time{}, false
	}
	to, err := time.ParseInLocation("2006-01-02", toDate, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_to"})
		return time.Time{}, time.Time{}, false
	}
	if to.Before(from) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_period"})
		return time.Time{}, time.Time{}, false
	}

	return from.UTC(), to.AddDate(0, 0, 1).UTC(), true
}
