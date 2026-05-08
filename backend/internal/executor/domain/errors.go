package domain

import "errors"

var ErrNilHTTPResponse = errors.New("http client returned nil response")
