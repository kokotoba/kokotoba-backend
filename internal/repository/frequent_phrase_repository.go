package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPhraseNotFound  = errors.New("frequent phrase not found")
	ErrPhraseDuplicate = errors.New("frequent phrase already exists")
)

type FrequentPhrase struct {
	ID   int64
	Text string
}

type FrequentPhraseRepository struct {
	database *pgxpool.Pool
}

func NewFrequentPhraseRepository(database *pgxpool.Pool) *FrequentPhraseRepository {
	return &FrequentPhraseRepository{database: database}
}

func (r *FrequentPhraseRepository) List(
	ctx context.Context,
	userID int64,
) ([]FrequentPhrase, error) {
	var userExists bool
	if err := r.database.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)",
		userID,
	).Scan(&userExists); err != nil {
		return nil, err
	}
	if !userExists {
		return nil, ErrUserNotFound
	}

	rows, err := r.database.Query(ctx, `
		SELECT id, text
		FROM frequent_phrases
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	phrases := make([]FrequentPhrase, 0)
	for rows.Next() {
		var phrase FrequentPhrase
		if err := rows.Scan(&phrase.ID, &phrase.Text); err != nil {
			return nil, err
		}
		phrases = append(phrases, phrase)
	}
	return phrases, rows.Err()
}

func (r *FrequentPhraseRepository) Create(
	ctx context.Context,
	userID int64,
	text string,
) (FrequentPhrase, error) {
	var phrase FrequentPhrase
	err := r.database.QueryRow(ctx, `
		INSERT INTO frequent_phrases (user_id, text)
		SELECT $1, $2
		WHERE EXISTS (SELECT 1 FROM users WHERE id = $1)
		RETURNING id, text
	`, userID, text).Scan(&phrase.ID, &phrase.Text)
	if errors.Is(err, pgx.ErrNoRows) {
		return FrequentPhrase{}, ErrUserNotFound
	}
	if isUniqueViolation(err) {
		return FrequentPhrase{}, ErrPhraseDuplicate
	}
	return phrase, err
}

func (r *FrequentPhraseRepository) Delete(
	ctx context.Context,
	userID int64,
	phraseID int64,
) error {
	result, err := r.database.Exec(ctx, `
		DELETE FROM frequent_phrases
		WHERE id = $1 AND user_id = $2
	`, phraseID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPhraseNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505"
}
