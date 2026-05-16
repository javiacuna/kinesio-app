package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/javiacuna/kinesio-backend/internal/finance/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListFinanciers(ctx context.Context, includeInactive bool) ([]domain.Financier, error) {
	var ms []FinancierModel
	q := r.db.WithContext(ctx).Order("name ASC")
	if !includeInactive {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Financier, 0, len(ms))
	for _, m := range ms {
		out = append(out, toFinancierDomain(m))
	}
	return out, nil
}

func (r *Repository) SaveFinancier(ctx context.Context, item domain.Financier) (domain.Financier, error) {
	m := FinancierModel{
		ID:     item.ID,
		Name:   item.Name,
		Kind:   item.Kind,
		Active: item.Active,
	}
	err := r.db.WithContext(ctx).Save(&m).Error
	return toFinancierDomain(m), err
}

func (r *Repository) ListTariffs(ctx context.Context, includeInactive bool) ([]domain.PracticeTariff, error) {
	var ms []PracticeTariffModel
	q := r.db.WithContext(ctx).Order("valid_from DESC, created_at DESC")
	if !includeInactive {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.PracticeTariff, 0, len(ms))
	for _, m := range ms {
		out = append(out, toTariffDomain(m))
	}
	return out, nil
}

func (r *Repository) SaveTariff(ctx context.Context, item domain.PracticeTariff) (domain.PracticeTariff, error) {
	m := PracticeTariffModel{
		ID:                item.ID,
		PracticeID:        item.PracticeID,
		FinancierID:       item.FinancierID,
		BillingValueCents: item.BillingValueCents,
		CopayCents:        item.CopayCents,
		Currency:          item.Currency,
		ValidFrom:         item.ValidFrom,
		ValidTo:           item.ValidTo,
		Active:            item.Active,
	}
	err := r.db.WithContext(ctx).Save(&m).Error
	return toTariffDomain(m), err
}

func (r *Repository) ListFeeRules(ctx context.Context, includeInactive bool) ([]domain.ProfessionalFeeRule, error) {
	var ms []ProfessionalFeeRuleModel
	q := r.db.WithContext(ctx).Order("created_at DESC")
	if !includeInactive {
		q = q.Where("active = ?", true)
	}
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ProfessionalFeeRule, 0, len(ms))
	for _, m := range ms {
		out = append(out, toFeeRuleDomain(m))
	}
	return out, nil
}

func (r *Repository) SaveFeeRule(ctx context.Context, item domain.ProfessionalFeeRule) (domain.ProfessionalFeeRule, error) {
	m := ProfessionalFeeRuleModel{
		ID:              item.ID,
		KinesiologistID: item.KinesiologistID,
		PracticeID:      item.PracticeID,
		RuleType:        item.RuleType,
		FixedValueCents: item.FixedValueCents,
		Percentage:      item.Percentage,
		Active:          item.Active,
	}
	err := r.db.WithContext(ctx).Save(&m).Error
	return toFeeRuleDomain(m), err
}

func (r *Repository) ListMovements(ctx context.Context, from, to *time.Time, status, kinesiologistID string) ([]domain.FinancialMovement, error) {
	var ms []FinancialMovementModel
	q := r.db.WithContext(ctx).Order("created_at DESC")
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if kinesiologistID != "" {
		q = q.Where("kinesiologist_id = ?", kinesiologistID)
	}
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.FinancialMovement, 0, len(ms))
	for _, m := range ms {
		out = append(out, toMovementDomain(m))
	}
	return out, nil
}

func (r *Repository) UpdateMovementStatuses(
	ctx context.Context,
	id uuid.UUID,
	collectionStatus *string,
	professionalPaymentStatus *string,
) (domain.FinancialMovement, error) {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}
	now := time.Now().UTC()
	if collectionStatus != nil {
		updates["collection_status"] = *collectionStatus
		if *collectionStatus == "collected" {
			updates["collected_at"] = now
		}
		if *collectionStatus == "pending" || *collectionStatus == "cancelled" {
			updates["collected_at"] = nil
		}
	}
	if professionalPaymentStatus != nil {
		updates["professional_payment_status"] = *professionalPaymentStatus
		if *professionalPaymentStatus == "paid" {
			updates["professional_paid_at"] = now
		}
		if *professionalPaymentStatus == "pending" || *professionalPaymentStatus == "cancelled" {
			updates["professional_paid_at"] = nil
		}
	}
	if err := r.db.WithContext(ctx).Model(&FinancialMovementModel{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return domain.FinancialMovement{}, err
	}
	var model FinancialMovementModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.FinancialMovement{}, domain.ErrNotFound
		}
		return domain.FinancialMovement{}, err
	}
	return toMovementDomain(model), nil
}

func (r *Repository) CompleteAppointment(ctx context.Context, appointmentID, practiceID, financierID uuid.UUID) (domain.FinancialMovement, error) {
	var movement domain.FinancialMovement
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing FinancialMovementModel
		err := tx.Where("appointment_id = ?", appointmentID).First(&existing).Error
		if err == nil {
			return domain.ErrAlreadyGenerated
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var appointment appointmentFinanceModel
		err = tx.Table("appointments").
			Select("id, patient_id, kinesiologist_id, practice_id, financier_id, status").
			Where("id = ?", appointmentID).
			First(&appointment).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if appointment.Status == "cancelled" {
			return domain.ErrInvalidStatus
		}
		if appointment.Status == "completed" {
			return domain.ErrAlreadyGenerated
		}

		var tariff PracticeTariffModel
		today := time.Now().UTC()
		err = tx.Where("practice_id = ? AND financier_id = ? AND active = ?", practiceID, financierID, true).
			Where("valid_from <= ?", today).
			Where("(valid_to IS NULL OR valid_to >= ?)", today).
			Order("valid_from DESC, created_at DESC").
			First(&tariff).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrTariffNotFound
			}
			return err
		}

		var feeRule ProfessionalFeeRuleModel
		var feeRuleID *uuid.UUID
		professionalFeeCents := int64(0)
		err = tx.Where("kinesiologist_id = ? AND practice_id = ? AND active = ?", appointment.KinesiologistID, practiceID, true).
			Order("created_at DESC").
			First(&feeRule).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			feeRuleID = &feeRule.ID
			if feeRule.RuleType == "fixed" && feeRule.FixedValueCents != nil {
				professionalFeeCents = *feeRule.FixedValueCents
			}
			if feeRule.RuleType == "percentage" && feeRule.Percentage != nil {
				professionalFeeCents = int64(float64(tariff.BillingValueCents)*(*feeRule.Percentage)/100 + 0.5)
			}
		}

		payerValueCents := tariff.BillingValueCents - tariff.CopayCents
		if payerValueCents < 0 {
			payerValueCents = 0
		}
		centerAmountCents := tariff.BillingValueCents - professionalFeeCents

		model := FinancialMovementModel{
			ID:                    uuid.New(),
			AppointmentID:         appointment.ID,
			PatientID:             appointment.PatientID,
			KinesiologistID:       appointment.KinesiologistID,
			PracticeID:            practiceID,
			FinancierID:           financierID,
			TariffID:              tariff.ID,
			FeeRuleID:             feeRuleID,
			BillingValueCents:     tariff.BillingValueCents,
			CopayCents:            tariff.CopayCents,
			PayerValueCents:       payerValueCents,
			ProfessionalFeeCents:  professionalFeeCents,
			CenterAmountCents:     centerAmountCents,
			Status:                "pending",
			CollectionStatus:      "pending",
			ProfessionalPayStatus: "pending",
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"status":       "completed",
			"practice_id":  practiceID,
			"financier_id": financierID,
			"updated_at":   time.Now().UTC(),
		}
		if err := tx.Table("appointments").Where("id = ?", appointment.ID).Updates(updates).Error; err != nil {
			return err
		}
		movement = toMovementDomain(model)
		return nil
	})
	return movement, err
}

func toFinancierDomain(m FinancierModel) domain.Financier {
	return domain.Financier{
		ID:        m.ID,
		Name:      m.Name,
		Kind:      m.Kind,
		Active:    m.Active,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toTariffDomain(m PracticeTariffModel) domain.PracticeTariff {
	return domain.PracticeTariff{
		ID:                m.ID,
		PracticeID:        m.PracticeID,
		FinancierID:       m.FinancierID,
		BillingValueCents: m.BillingValueCents,
		CopayCents:        m.CopayCents,
		Currency:          m.Currency,
		ValidFrom:         m.ValidFrom,
		ValidTo:           m.ValidTo,
		Active:            m.Active,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func toFeeRuleDomain(m ProfessionalFeeRuleModel) domain.ProfessionalFeeRule {
	return domain.ProfessionalFeeRule{
		ID:              m.ID,
		KinesiologistID: m.KinesiologistID,
		PracticeID:      m.PracticeID,
		RuleType:        m.RuleType,
		FixedValueCents: m.FixedValueCents,
		Percentage:      m.Percentage,
		Active:          m.Active,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toMovementDomain(m FinancialMovementModel) domain.FinancialMovement {
	return domain.FinancialMovement{
		ID:                    m.ID,
		AppointmentID:         m.AppointmentID,
		PatientID:             m.PatientID,
		KinesiologistID:       m.KinesiologistID,
		PracticeID:            m.PracticeID,
		FinancierID:           m.FinancierID,
		TariffID:              m.TariffID,
		FeeRuleID:             m.FeeRuleID,
		BillingValueCents:     m.BillingValueCents,
		CopayCents:            m.CopayCents,
		PayerValueCents:       m.PayerValueCents,
		ProfessionalFeeCents:  m.ProfessionalFeeCents,
		CenterAmountCents:     m.CenterAmountCents,
		Status:                m.Status,
		CollectionStatus:      m.CollectionStatus,
		ProfessionalPayStatus: m.ProfessionalPayStatus,
		CollectedAt:           m.CollectedAt,
		ProfessionalPaidAt:    m.ProfessionalPaidAt,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}
