package router

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"kokotoba-backend/internal/controller"
	"kokotoba-backend/internal/repository"
)

type Middleware func(http.Handler) http.Handler

func NewMux(database *pgxpool.Pool, authenticate Middleware) *http.ServeMux {
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
	handleAuthenticated(mux, "GET /api/v1", serviceController.Show, authenticate)
	handleAuthenticated(mux, "GET /api/v1/me/settings", userSettingsController.Show, authenticate)
	handleAuthenticated(mux, "PATCH /api/v1/me/settings", userSettingsController.Update, authenticate)
	handleAuthenticated(mux, "GET /api/v1/me/phrases", frequentPhraseController.Index, authenticate)
	handleAuthenticated(mux, "POST /api/v1/me/phrases", frequentPhraseController.Create, authenticate)
	handleAuthenticated(mux, "PUT /api/v1/me/phrases/order", frequentPhraseController.Reorder, authenticate)
	handleAuthenticated(
		mux,
		"DELETE /api/v1/me/phrases/{phraseID}",
		frequentPhraseController.Delete,
		authenticate,
	)
	return mux
}

func handleAuthenticated(
	mux *http.ServeMux,
	pattern string,
	handler http.HandlerFunc,
	authenticate Middleware,
) {
	mux.Handle(pattern, authenticate(handler))
}
