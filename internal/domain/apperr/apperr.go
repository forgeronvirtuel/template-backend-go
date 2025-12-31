package apperr

import "fmt"

type Code string

const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeConflict        Code = "CONFLICT"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeValidationError Code = "VALIDATION_ERROR"
	CodeInternal        Code = "INTERNAL_ERROR"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func Wrap(code Code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

func New(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}
