package domain

import "errors"

var (
	ErrValidation            = errors.New("validation error")
	ErrOverlap               = errors.New("overlap")
	ErrNotFound              = errors.New("not found")
	ErrPatientNotFound       = errors.New("patient not found")
	ErrKinesiologistNotFound = errors.New("kinesiologist not found")
	ErrInvalidStatus         = errors.New("invalid status")
)
