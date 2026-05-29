package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"service":   "habit",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
