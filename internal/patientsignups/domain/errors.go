package domain

import "errors"

var (
	ErrValidation      = errors.New("validation error")
	ErrEmailInUse      = errors.New("email already registered")
	ErrAlreadyPending  = errors.New("signup already pending")
	ErrRequestNotFound = errors.New("request not found")
	ErrAlreadyReviewed = errors.New("already reviewed")
	ErrPatientConflict = errors.New("patient conflict")
)
