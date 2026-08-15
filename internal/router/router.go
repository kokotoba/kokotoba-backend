package router

import (
	"context"
	"net/http"

	"kokotoba-backend/internal/controller"
)

type databasePinger interface {
	Ping(context.Context) error
}

func NewMux(database databasePinger) *http.ServeMux {
	mux := http.NewServeMux()
	healthController := controller.NewHealthController(database)
	serviceController := controller.NewServiceController("kokotoba-backend", "dev")

	mux.HandleFunc("GET /healthz", healthController.Show)
	mux.HandleFunc("GET /api/v1", serviceController.Show)
	return mux
}
