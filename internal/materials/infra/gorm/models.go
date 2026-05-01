package gorm

import "time"

type MaterialModel struct {
	ID             string `gorm:"type:uuid;primaryKey"`
	Name           string `gorm:"not null"`
	Description    *string
	TotalQty       int `gorm:"not null"`
	AvailableQty   int `gorm:"not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CreatedByEmail *string
	CreatedByRole  *string
	UpdatedByEmail *string
	UpdatedByRole  *string
}

func (MaterialModel) TableName() string { return "materials" }

type MaterialLoanModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	MaterialID      string `gorm:"type:uuid;not null"`
	PatientID       string `gorm:"type:uuid;not null"`
	KinesiologistID string `gorm:"type:uuid;not null"`
	Qty             int    `gorm:"not null"`
	Notes           *string
	LoanedAt        time.Time
	ReturnedAt      *time.Time
	LoanedByEmail   *string
	LoanedByRole    *string
	ReturnedByEmail *string
	ReturnedByRole  *string
}

func (MaterialLoanModel) TableName() string { return "material_loans" }
