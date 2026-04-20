package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard error codes
const (
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeRateLimited  = "RATE_LIMITED"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError sends a standardized error response
func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": ErrorResponse{
			Code:    code,
			Message: message,
		},
	})
}

// respondSuccess sends a standardized success response with data
func respondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"data": data,
	})
}

// Common error responses

func respondValidationError(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, ErrCodeValidation, message)
}

func respondUnauthorized(c *gin.Context, message string) {
	respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

func respondForbidden(c *gin.Context, message string) {
	respondError(c, http.StatusForbidden, ErrCodeForbidden, message)
}

func respondNotFound(c *gin.Context, message string) {
	respondError(c, http.StatusNotFound, ErrCodeNotFound, message)
}

func respondConflict(c *gin.Context, message string) {
	respondError(c, http.StatusConflict, ErrCodeConflict, message)
}

func respondInternalError(c *gin.Context, message string) {
	respondError(c, http.StatusInternalServerError, ErrCodeInternal, message)
}

func respondBadRequest(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, ErrCodeBadRequest, message)
}

func respondRateLimited(c *gin.Context) {
	respondError(c, http.StatusTooManyRequests, ErrCodeRateLimited, "Too many requests, please try again later")
}

// Pagination types

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// respondPaginated sends a paginated response
func respondPaginated(c *gin.Context, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": meta,
	})
}

// Success responses

func respondOK(c *gin.Context, data interface{}) {
	respondSuccess(c, http.StatusOK, data)
}

func respondCreated(c *gin.Context, data interface{}) {
	respondSuccess(c, http.StatusCreated, data)
}

func respondMessage(c *gin.Context, message string) {
	respondSuccess(c, http.StatusOK, gin.H{"message": message})
}
