package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	codeValidation = "VALIDATION_ERROR"
	codeBadRequest = "BAD_REQUEST"
	codeNotFound   = "NOT_FOUND"
	codeInternal   = "INTERNAL_ERROR"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondError(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": errorBody{Code: code, Message: msg}})
}

func respondData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func respondOK(c *gin.Context, data any)      { respondData(c, http.StatusOK, data) }
func respondCreated(c *gin.Context, data any) { respondData(c, http.StatusCreated, data) }

func respondMessage(c *gin.Context, msg string) {
	respondData(c, http.StatusOK, gin.H{"message": msg})
}
