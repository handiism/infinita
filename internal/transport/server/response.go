package server

import (
	"encoding/json"
	"errors"
	"net/http"

	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

type Response struct {
	Status  Status      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
	Meta    interface{} `json:"meta"`
}

type ErrorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

func Success(data interface{}) Response {
	return Response{
		Status: StatusSuccess,
		Data:   data,
		Meta:   nil,
	}
}

func SuccessWithMeta(data interface{}, meta interface{}) Response {
	return Response{
		Status: StatusSuccess,
		Data:   data,
		Meta:   meta,
	}
}

func Fail(errors []domainerror.DomainError) Response {
	errorObjects := make([]ErrorObject, len(errors))
	for i, e := range errors {
		errorObjects[i] = ErrorObject{
			Code:    e.Code,
			Message: e.Message,
			Field:   e.Field,
			Hint:    e.Hint,
		}
	}
	return Response{
		Status: StatusFail,
		Data:   errorObjects,
		Meta:   nil,
	}
}

func FailFromSingle(err domainerror.DomainError) Response {
	return Fail([]domainerror.DomainError{err})
}

func Error(code, message string) Response {
	return Response{
		Status:  StatusError,
		Code:    code,
		Message: message,
		Meta:    nil,
	}
}

func WriteJSON(w http.ResponseWriter, statusCode int, resp Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(resp)
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data interface{}) error {
	return WriteJSON(w, statusCode, Success(data))
}

func WriteSuccessWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta interface{}) error {
	return WriteJSON(w, statusCode, SuccessWithMeta(data, meta))
}

func WriteFail(w http.ResponseWriter, statusCode int, errors []domainerror.DomainError) error {
	return WriteJSON(w, statusCode, Fail(errors))
}

func WriteFailFromError(w http.ResponseWriter, statusCode int, err error) bool {
	var domainErr domainerror.DomainError
	if !errors.As(err, &domainErr) {
		return false
	}

	_ = WriteFail(w, statusCode, []domainerror.DomainError{domainErr})
	return true
}

func WriteError(w http.ResponseWriter, statusCode int, code, message string) error {
	return WriteJSON(w, statusCode, Error(code, message))
}
