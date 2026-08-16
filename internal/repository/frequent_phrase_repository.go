package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPhraseNotFound     = errors.New("frequent phrase not found")
	ErrPhraseDuplicate    = errors.New("frequent phrase already exists")
	ErrPhraseOrderInvalid = errors.New("frequent phrase order is invalid")
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
		ORDER BY sort_order, id
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
	tx, err := r.database.Begin(ctx)
	if err != nil {
		return FrequentPhrase{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockUser(ctx, tx, userID); err != nil {
		return FrequentPhrase{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE frequent_phrases
		SET sort_order = sort_order + 1
		WHERE user_id = $1
	`, userID); err != nil {
		return FrequentPhrase{}, err
	}

	var phrase FrequentPhrase
	err = tx.QueryRow(ctx, `
		INSERT INTO frequent_phrases (user_id, text, sort_order)
		VALUES ($1, $2, 0)
		RETURNING id, text
	`, userID, text).Scan(&phrase.ID, &phrase.Text)
	if isUniqueViolation(err) {
		return FrequentPhrase{}, ErrPhraseDuplicate
	}
	if err != nil {
		return FrequentPhrase{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FrequentPhrase{}, err
	}
	return phrase, nil
}

func (r *FrequentPhraseRepository) Delete(
	ctx context.Context,
	userID int64,
	phraseID int64,
) error {
	tx, err := r.database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockUser(ctx, tx, userID); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `
		DELETE FROM frequent_phrases
		WHERE id = $1 AND user_id = $2
	`, phraseID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPhraseNotFound
	}
	return tx.Commit(ctx)
}

func (r *FrequentPhraseRepository) Reorder(
	ctx context.Context,
	userID int64,
	phraseIDs []int64,
) error {
	tx, err := r.database.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := lockUser(ctx, tx, userID); err != nil {
		return err
	}

	var phraseCount int
	if err := tx.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM frequent_phrases WHERE user_id = $1",
		userID,
	).Scan(&phraseCount); err != nil {
		return err
	}
	if phraseCount != len(phraseIDs) {
		return ErrPhraseOrderInvalid
	}

	result, err := tx.Exec(ctx, `
		UPDATE frequent_phrases AS phrase
		SET
			sort_order = requested.position - 1,
			updated_at = CURRENT_TIMESTAMP
		FROM UNNEST($2::BIGINT[]) WITH ORDINALITY AS requested(id, position)
		WHERE phrase.user_id = $1 AND phrase.id = requested.id
	`, userID, phraseIDs)
	if err != nil {
		return err
	}
	if result.RowsAffected() != int64(len(phraseIDs)) {
		return ErrPhraseOrderInvalid
	}
	return tx.Commit(ctx)
}

type userLocker interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func lockUser(ctx context.Context, database userLocker, userID int64) error {
	var lockedUserID int64
	err := database.QueryRow(
		ctx,
		"SELECT id FROM users WHERE id = $1 FOR UPDATE",
		userID,
	).Scan(&lockedUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505"
}
