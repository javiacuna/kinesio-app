package domain

import "errors"

var (
	ErrValidation       = errors.New("validation_error")
	ErrNotFound         = errors.New("not_found")
	ErrTariffNotFound   = errors.New("tariff_not_found")
	ErrAlreadyGenerated = errors.New("financial_movement_already_generated")
	ErrInvalidStatus    = errors.New("invalid_appointment_status")
)
