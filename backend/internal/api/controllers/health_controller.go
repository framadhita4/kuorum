package controllers

import (
	"net/http"
	"time"

	"kuorum/backend/internal/api/response"
)

type HealthController struct {
	startTime time.Time
}

func NewHealthController() *HealthController {
	return &HealthController{
		startTime: time.Now().UTC(),
	}
}

func (c *HealthController) Check(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, "System operational", map[string]any{
		"status":    "healthy",
		"uptime":    time.Since(c.startTime).String(),
		"timestamp": time.Now().UTC(),
		"service":   "kuorum-api",
	})
}
