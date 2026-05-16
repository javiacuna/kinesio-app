package gorm

import (
	"time"

	"github.com/google/uuid"
)

type FinancierModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	Name      string    `gorm:"column:name"`
	Kind      string    `gorm:"column:kind"`
	Active    bool      `gorm:"column:active"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (FinancierModel) TableName() string { return "financiers" }

type PracticeTariffModel struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	PracticeID        uuid.UUID  `gorm:"type:uuid;column:practice_id"`
	FinancierID       uuid.UUID  `gorm:"type:uuid;column:financier_id"`
	BillingValueCents int64      `gorm:"column:billing_value_cents"`
	CopayCents        int64      `gorm:"column:copay_cents"`
	Currency          string     `gorm:"column:currency"`
	ValidFrom         time.Time  `gorm:"column:valid_from"`
	ValidTo           *time.Time `gorm:"column:valid_to"`
	Active            bool       `gorm:"column:active"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (PracticeTariffModel) TableName() string { return "practice_tariffs" }

type ProfessionalFeeRuleModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	KinesiologistID uuid.UUID `gorm:"type:uuid;column:kinesiologist_id"`
	PracticeID      uuid.UUID `gorm:"type:uuid;column:practice_id"`
	RuleType        string    `gorm:"column:rule_type"`
	FixedValueCents *int64    `gorm:"column:fixed_value_cents"`
	Percentage      *float64  `gorm:"column:percentage"`
	Active          bool      `gorm:"column:active"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProfessionalFeeRuleModel) TableName() string { return "professional_fee_rules" }

type FinancialMovementModel struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;column:id"`
	AppointmentID         uuid.UUID  `gorm:"type:uuid;column:appointment_id"`
	PatientID             uuid.UUID  `gorm:"type:uuid;column:patient_id"`
	KinesiologistID       uuid.UUID  `gorm:"type:uuid;column:kinesiologist_id"`
	PracticeID            uuid.UUID  `gorm:"type:uuid;column:practice_id"`
	FinancierID           uuid.UUID  `gorm:"type:uuid;column:financier_id"`
	TariffID              uuid.UUID  `gorm:"type:uuid;column:tariff_id"`
	FeeRuleID             *uuid.UUID `gorm:"type:uuid;column:fee_rule_id"`
	BillingValueCents     int64      `gorm:"column:billing_value_cents"`
	CopayCents            int64      `gorm:"column:copay_cents"`
	PayerValueCents       int64      `gorm:"column:payer_value_cents"`
	ProfessionalFeeCents  int64      `gorm:"column:professional_fee_cents"`
	CenterAmountCents     int64      `gorm:"column:center_amount_cents"`
	Status                string     `gorm:"column:status"`
	CollectionStatus      string     `gorm:"column:collection_status"`
	ProfessionalPayStatus string     `gorm:"column:professional_payment_status"`
	CollectedAt           *time.Time `gorm:"column:collected_at"`
	ProfessionalPaidAt    *time.Time `gorm:"column:professional_paid_at"`
	CancellationReason    *string    `gorm:"column:cancellation_reason"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (FinancialMovementModel) TableName() string { return "financial_movements" }

type appointmentFinanceModel struct {
	ID              uuid.UUID  `gorm:"column:id"`
	PatientID       uuid.UUID  `gorm:"column:patient_id"`
	KinesiologistID uuid.UUID  `gorm:"column:kinesiologist_id"`
	PracticeID      *uuid.UUID `gorm:"column:practice_id"`
	FinancierID     *uuid.UUID `gorm:"column:financier_id"`
	Status          string     `gorm:"column:status"`
}
