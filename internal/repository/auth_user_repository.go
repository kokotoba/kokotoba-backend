package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultUserName = "ユーザー"

type AuthUserRepository struct {
	database *pgxpool.Pool
}

func NewAuthUserRepository(database *pgxpool.Pool) *AuthUserRepository {
	return &AuthUserRepository{database: database}
}

func (r *AuthUserRepository) ResolveFirebaseUser(
	ctx context.Context,
	firebaseUID string,
	displayName string,
) (int64, error) {
	var userID int64
	err := r.database.QueryRow(
		ctx,
		"SELECT id FROM users WHERE firebase_uid = $1",
		firebaseUID,
	).Scan(&userID)
	if err == nil {
		return userID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	name := strings.TrimSpace(displayName)
	if name == "" {
		name = defaultUserName
	}

	err = r.database.QueryRow(ctx, `
		INSERT INTO users (name, firebase_uid)
		VALUES ($1, $2)
		ON CONFLICT (firebase_uid) DO UPDATE
		SET firebase_uid = EXCLUDED.firebase_uid
		RETURNING id
	`, name, firebaseUID).Scan(&userID)
	return userID, err
}

func (r *AuthUserRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	var foundUserID int64
	err := r.database.QueryRow(ctx, "SELECT id FROM users WHERE id = $1", userID).Scan(&foundUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
