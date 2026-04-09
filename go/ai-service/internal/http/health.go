package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadyCheck is a single dependency probe. Returns nil when healthy.
type ReadyCheck func() error

// RegisterHealthRoutes adds GET /health (liveness) and GET /ready (dependency probes).
func RegisterHealthRoutes(r *gin.Engine, checks map[string]ReadyCheck) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		results := map[string]string{}
		allOK := true
		for name, fn := range checks {
			if err := fn(); err != nil {
				results[name] = err.Error()
				allOK = false
			} else {
				results[name] = "ok"
			}
		}
		status := http.StatusOK
		if !allOK {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"checks": results})
	})
}
