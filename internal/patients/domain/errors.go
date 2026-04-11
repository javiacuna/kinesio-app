package domain

import "errors"

var (
	ErrDuplicateDNI   = errors.New("duplicate dni")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrNotFound       = errors.New("not found")
	ErrValidation     = errors.New("validation error")
)
