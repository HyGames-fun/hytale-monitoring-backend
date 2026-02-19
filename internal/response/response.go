package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ValidationErrorResponse struct {
	Status string              `json:"status"`
	Errors map[string][]string `json:"errors"`
}

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

func Error(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Status:  "error",
		Message: message,
		Code:    code,
	})
}

func ValidationError(w http.ResponseWriter, errors map[string][]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(ValidationErrorResponse{
		Status: "validation_error",
		Errors: errors,
	})
}
