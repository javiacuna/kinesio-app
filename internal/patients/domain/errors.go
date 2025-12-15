package domain

import "errors"

var (
	ErrDuplicateDNI   = errors.New("duplicate dni")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrValidation     = errors.New("validation error")
)
