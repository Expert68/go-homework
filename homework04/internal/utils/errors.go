package utils

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	CodeOK           ErrorCode = 0
	CodeInvalidParam ErrorCode = 40000
	CodeUnauthorized ErrorCode = 40100
	CodeForbidden    ErrorCode = 40300
	CodeNotFound     ErrorCode = 40400
	CodeConflict     ErrorCode = 40900
	CodeInternal     ErrorCode = 50000
)

func (c ErrorCode) HTTPStatus() int {
	switch c {
	case CodeOK:
		return http.StatusOK
	case CodeInvalidParam:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"error"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error              { return e.Err }
func New(code ErrorCode, msg string) *AppError { return &AppError{Code: code, Message: msg} }
func Wrap(code ErrorCode, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

func BadRequest(msg string) *AppError   { return New(CodeInvalidParam, msg) }
func Unauthorized(msg string) *AppError { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *AppError    { return New(CodeForbidden, msg) }
func NotFound(msg string) *AppError     { return New(CodeNotFound, msg) }
func Conflict(msg string) *AppError     { return New(CodeConflict, msg) }
func Internal(err error) *AppError      { return Wrap(CodeInternal, "服务器内部错误", err) }

func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
