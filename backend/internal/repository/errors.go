package repository

import "errors"

// Sentinel errors for repository layer.
var (
	ErrNotFound   = errors.New("record not found")
	ErrDuplicate  = errors.New("duplicate record")
	ErrConstraint = errors.New("constraint violation")
)
