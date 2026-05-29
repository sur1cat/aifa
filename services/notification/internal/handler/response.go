package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codeValidation = "VALIDATION_ERROR"
	codeInternal   = "INTERNAL_ERROR"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": errorBody{Code: code, Message: msg}})
}

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}
