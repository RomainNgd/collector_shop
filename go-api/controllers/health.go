package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pinger is satisfied by *database.Database; kept minimal so this package
// does not need to import database and tests can supply a stub.
type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Healthz is the liveness probe: it only reports that the process is up and
// serving, without touching any dependency, so a slow database never causes
// Kubernetes to kill and restart an otherwise healthy pod.
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz is the readiness probe: it checks the database connection so
// Kubernetes stops routing traffic to a pod that cannot serve requests.
func (h *HealthHandler) Readyz(c *gin.Context) {
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
