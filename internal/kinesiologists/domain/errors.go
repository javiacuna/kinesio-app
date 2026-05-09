package domain

import "errors"

var (
	ErrValidation     = errors.New("validation error")
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrDuplicateName  = errors.New("duplicate name")
)
