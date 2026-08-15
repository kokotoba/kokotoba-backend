package controller

import (
	"context"
	"net/http"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/view"
)

type databasePinger interface {
	Ping(context.Context) error
}

type HealthController struct {
	database databasePinger
}

func NewHealthController(database databasePinger) *HealthController {
	return &HealthController{database: database}
}

func (c *HealthController) Show(w http.ResponseWriter, r *http.Request) {
	if err := c.database.Ping(r.Context()); err != nil {
		_ = view.JSON(w, http.StatusServiceUnavailable, model.HealthResponse{
			Status:   "unavailable",
			Database: "down",
		})
		return
	}
	_ = view.JSON(w, http.StatusOK, model.HealthResponse{
		Status:   "ok",
		Database: "ok",
	})
}
