package domain

import "errors"

var (
	ErrValidation        = errors.New("validation_error")
	ErrNotFound          = errors.New("not_found")
	ErrDuplicateName     = errors.New("duplicate_name")
	ErrInsufficientStock = errors.New("insufficient_stock")
	ErrAlreadyReturned   = errors.New("already_returned")
)
