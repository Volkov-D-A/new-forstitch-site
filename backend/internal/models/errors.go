package models

import "errors"

var (
	ErrBadRequest   = errors.New("bad request")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation error")
)

type AppError struct {
	Kind    error
	Code    string
	Message string
}

func (err AppError) Error() string {
	return err.Message
}

func (err AppError) Is(target error) bool {
	return target == err.Kind
}

type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func BadRequest(code string, message string) error {
	return AppError{Kind: ErrBadRequest, Code: code, Message: message}
}

func Conflict(code string, message string) error {
	return AppError{Kind: ErrConflict, Code: code, Message: message}
}

func Internal(code string, message string) error {
	return AppError{Kind: ErrInternal, Code: code, Message: message}
}

func NotFound(code string, message string) error {
	return AppError{Kind: ErrNotFound, Code: code, Message: message}
}

func Unauthorized(code string, message string) error {
	return AppError{Kind: ErrUnauthorized, Code: code, Message: message}
}

func Validation(code string, message string) error {
	return AppError{Kind: ErrValidation, Code: code, Message: message}
}

func ErrorPayloadFrom(err error, fallbackCode string, fallbackMessage string) ErrorPayload {
	var appErr AppError
	if errors.As(err, &appErr) {
		return ErrorPayload{Code: appErr.Code, Message: appErr.Message}
	}
	return ErrorPayload{Code: fallbackCode, Message: fallbackMessage}
}
