package http

import "net/http"

type StatusError struct {
	statusCode int
	response   *http.Response
	message    string
}

func (e *StatusError) Error() string {
	return e.message
}

func NewHTTPStatusError(statusCode int, response *http.Response, message string) error {
	return &StatusError{
		statusCode: statusCode,
		response:   response,
		message:    message,
	}
}
