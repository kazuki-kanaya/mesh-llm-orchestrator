package domain

import "errors"

var (
	ErrJobNotFound     = errors.New("job not found")
	ErrNilHTTPResponse = errors.New("http response is nil")
)
