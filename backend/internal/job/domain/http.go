package domain

import "net/http"

type HTTPRequest struct {
	Method  string
	URL     string
	Host    string
	Headers http.Header
	Body    []byte
}

type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}
