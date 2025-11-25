package response

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Pagination `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type Pagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Success sends successful response
func Success(w http.ResponseWriter, statusCode int, data interface{}, pagination *Pagination) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success:   true,
		Data:      data,
		Meta:      pagination,
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// Error sends error response
func Error(w http.ResponseWriter, statusCode int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// BadRequest for validation errors
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "VALIDATION_ERROR", message, "")
}

// NotFound for resource not found
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message, "")
}

// Unauthorized for auth errors
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message, "")
}

// InternalServerError for server errors
func InternalServerError(w http.ResponseWriter, details string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", details)
}

// Conflict for duplicate/conflict errors
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message, "")
}

// Created for successful creation
func Created(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusCreated, data, nil)
}

// OK for successful operations
func OK(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusOK, data, nil)
}

// OKPaged for paginated data
func OKPaged(w http.ResponseWriter, data interface{}, page, size, total int) {
	pagination := &Pagination{Page: page, Size: size, Total: total}
	Success(w, http.StatusOK, data, pagination)
}

// NoContent for successful operations with no content
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
