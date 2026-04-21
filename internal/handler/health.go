package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
	Database  string `json:"database,omitempty"`
}

func HealthWithDB(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		stat := pool.Stat()

		dbStatus := "ok"
		httpStatus := http.StatusOK
		overallStatus := "ok"

		// Pool has no idle connections and is fully acquired — DB likely unhealthy
		if stat.TotalConns() > 0 && stat.IdleConns() == 0 && stat.AcquiredConns() >= stat.MaxConns() {
			dbStatus = "unavailable"
			httpStatus = http.StatusServiceUnavailable
			overallStatus = "degraded"
		}

		c.JSON(httpStatus, HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Version:   "1.0.0",
			Database:  dbStatus,
		})
	}
}
