package domain

import "errors"

var (
	ErrNotFound      = errors.New("product not found")
	ErrAlreadyExists = errors.New("product already exists")
	ErrNoTx          = errors.New("transaction is required")
)
