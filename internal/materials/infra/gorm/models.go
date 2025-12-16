package gorm

import "time"

type MaterialModel struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Name         string `gorm:"not null"`
	Description  *string
	TotalQty     int `gorm:"not null"`
	AvailableQty int `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
}

func (MaterialLoanModel) TableName() string { return "material_loans" }
