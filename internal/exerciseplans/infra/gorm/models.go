package gorm

import "time"

type ExercisePlanModel struct {
	ID              string `gorm:"type:uuid;primaryKey"`
	PatientID       string `gorm:"type:uuid;not null"`
	KinesiologistID string `gorm:"type:uuid;not null"`
	Frequency       string `gorm:"not null"`
	DurationWeeks   int    `gorm:"not null"`
	Observations    *string
	Status          string `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Items []ExercisePlanItemModel `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE"`
}

func (ExercisePlanModel) TableName() string { return "exercise_plans" }

type ExercisePlanItemModel struct {
	ID               string `gorm:"type:uuid;primaryKey"`
	PlanID           string `gorm:"type:uuid;not null"`
	Name             string `gorm:"not null"`
	Description      *string
	VideoURL         *string
	GuideURL         *string
	EstimatedMinutes int `gorm:"not null"`
	Sets             *int
	Reps             *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (ExercisePlanItemModel) TableName() string { return "exercise_plan_items" }
