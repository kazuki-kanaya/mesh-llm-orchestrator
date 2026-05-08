package domain

import "errors"

var (
	ErrInvalidStaleAfter = errors.New("stale after must be positive")
	ErrInvalidBatchSize  = errors.New("batch size must be positive")
)
