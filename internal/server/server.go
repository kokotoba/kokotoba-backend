package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"kokotoba-backend/internal/database"
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

	srv := &http.Server{
		Addr:    addr,
		Handler: router.NewMux(db),
	}

	log.Printf("kokotoba-backend listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
