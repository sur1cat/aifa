package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

func paginationParams(c *gin.Context) (int, int) {
	limit, offset := defaultListLimit, 0
	if s := c.Query("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= maxListLimit {
			limit = v
		}
	}
	if s := c.Query("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}
