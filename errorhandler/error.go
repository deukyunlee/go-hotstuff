package errorhandler

import (
	"fmt"
	"net/http"
)

type ErrorLogLevel string

const (
	ERROR ErrorLogLevel = "ERROR"
)

type CustomErrorType struct {
	Code       int
	HTTPStatus int
	LogLevel   ErrorLogLevel
}

const (
	StatusOK                  = http.StatusOK
	StatusBadRequest          = http.StatusBadRequest
	StatusInternalServerError = http.StatusInternalServerError
	StatusNotFound            = http.StatusNotFound
	StatusTooManyRequests     = http.StatusTooManyRequests
	StatusUnprocessableEntity = http.StatusUnprocessableEntity
)

type CustomError struct {
	ErrType CustomErrorType
	Err     error
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("Code: %d, Status: %d, Level: %s, Message: %s",
		e.ErrType.Code, e.ErrType.HTTPStatus, e.ErrType.LogLevel, e.Err.Error())
}

func LogError(customError CustomErrorType, err error) {
	fmt.Printf("Code: %d, Status: %d, Level: %s, Message: %s\n",
		customError.Code, customError.HTTPStatus, customError.LogLevel, err.Error())
}

var (
	HTTP_WRITE_CONTENT_ERROR = CustomErrorType{1000, StatusInternalServerError, ERROR}
)
