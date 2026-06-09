package domain

import "errors"

var (
	ErrValidation            = errors.New("validation error")
	ErrOverlap               = errors.New("overlap")
	ErrNotFound              = errors.New("not found")
	ErrPatientNotFound       = errors.New("patient not found")
	ErrPatientInactive       = errors.New("patient inactive")
	ErrKinesiologistNotFound = errors.New("kinesiologist not found")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrVideoProviderMissing  = errors.New("video provider missing")
	ErrVideoProviderFailed   = errors.New("video provider failed")
)
