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
	frequentPhraseController := controller.NewFrequentPhraseController(
		repository.NewFrequentPhraseRepository(database),
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
	mux.HandleFunc(
		"GET /api/v1/users/{userID}/phrases",
		frequentPhraseController.Index,
	)
	mux.HandleFunc(
		"POST /api/v1/users/{userID}/phrases",
		frequentPhraseController.Create,
	)
	mux.HandleFunc(
		"PUT /api/v1/users/{userID}/phrases/order",
		frequentPhraseController.Reorder,
	)
	mux.HandleFunc(
		"DELETE /api/v1/users/{userID}/phrases/{phraseID}",
		frequentPhraseController.Delete,
	)
	return mux
}
