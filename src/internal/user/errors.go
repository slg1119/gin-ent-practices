package user

import "errors"

var (
	ErrInvalidName       = errors.New("name must contain 1 to 100 characters")
	ErrInvalidEmail      = errors.New("email must be a valid email address")
	ErrInvalidID         = errors.New("user ID must be a positive integer")
	ErrInvalidPagination = errors.New("limit must be 1 to 100 and offset must be non-negative")
	ErrNotFound          = errors.New("user not found")
	ErrEmailTaken        = errors.New("email is already registered")
)
