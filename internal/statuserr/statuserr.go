package statuserr

import (
	"errors"
	"net/http"
)

type Error struct {
	code    int
	message string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() int {
	return e.code
}

func New(code int, message string) error {
	return &Error{
		code:    code,
		message: message,
	}
}

func StatusCode(err error) (int, bool) {
	type statusCoder interface {
		StatusCode() int
	}

	if err == nil {
		return 0, false
	}

	var statusErr statusCoder
	if errors.As(err, &statusErr) {
		return statusErr.StatusCode(), true
	}

	return 0, false
}

func NotFound(message string) error {
	return New(http.StatusNotFound, message)
}

func Conflict(message string) error {
	return New(http.StatusConflict, message)
}

func Forbidden(message string) error {
	return New(http.StatusForbidden, message)
}

func Unauthorized(message string) error {
	return New(http.StatusUnauthorized, message)
}

func Internal(message string) error {
	return New(http.StatusInternalServerError, message)
}

func ServiceUnavailable(message string) error {
	return New(http.StatusServiceUnavailable, message)
}
