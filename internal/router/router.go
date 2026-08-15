package router

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"kokotoba-backend/internal/controller"
	"kokotoba-backend/internal/repository"
)

func NewMux(database *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()
	healthController := controller.NewHealthController(database)
	serviceController := controller.NewServiceController("kokotoba-backend", "dev")
	userSettingsController := controller.NewUserSettingsController(
		repository.NewUserRepository(database),
	)

	mux.HandleFunc("GET /healthz", healthController.Show)
	mux.HandleFunc("GET /api/v1", serviceController.Show)
	mux.HandleFunc(
		"GET /api/v1/users/{userID}/settings",
		userSettingsController.Show,
	)
	mux.HandleFunc(
		"PATCH /api/v1/users/{userID}/settings",
		userSettingsController.Update,
	)
	return mux
}
