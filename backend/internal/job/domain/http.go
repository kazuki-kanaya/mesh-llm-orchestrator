package domain

import "net/http"

type HTTPRequest struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
