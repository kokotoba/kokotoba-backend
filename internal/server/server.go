package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"kokotoba-backend/internal/authn"
	"kokotoba-backend/internal/database"
	"kokotoba-backend/internal/repository"
	"kokotoba-backend/internal/router"
)

func Run(addr string) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://kokotoba:kokotoba_dev_password@localhost:5432/kokotoba"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	authConfig, err := authn.LoadConfig(os.Getenv)
	if err != nil {
		return fmt.Errorf("load authentication config: %w", err)
	}
	authMiddleware, err := initializeAuthentication(ctx, db, authConfig)
	if err != nil {
		return fmt.Errorf("initialize authentication: %w", err)
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: router.WithCORS(router.NewMux(db, authMiddleware.Authenticate)),
	}

	log.Printf("kokotoba-backend listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func initializeAuthentication(
	ctx context.Context,
	db *pgxpool.Pool,
	config authn.Config,
) (*authn.Middleware, error) {
	authUsers := repository.NewAuthUserRepository(db)
	if config.Mode == authn.ModeDev {
		exists, err := authUsers.UserExists(ctx, config.DevUserID)
		if err != nil {
			return nil, fmt.Errorf("validate DEV_USER_ID: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("DEV_USER_ID %d does not exist", config.DevUserID)
		}
		return authn.NewDevMiddleware(config.DevUserID), nil
	}

	verifier, err := authn.NewFirebaseTokenVerifier(ctx, config.FirebaseProjectID)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase: %w", err)
	}
	return authn.NewFirebaseMiddleware(verifier, authUsers), nil
}
