package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConversationSession struct {
	ID        int64
	StartedAt time.Time
	EndedAt   *time.Time
}

type Utterance struct {
	ID       int64
	Speaker  string
	Text     string
	SpokenAt time.Time
}

var ErrSessionNotFound = errors.New("conversation session not found")

type ConversationSessionRepository struct {
	database *pgxpool.Pool
}

func NewConversationSessionRepository(
	database *pgxpool.Pool,
) *ConversationSessionRepository {
	return &ConversationSessionRepository{database: database}
}

// StartOrResume は進行中のセッションがあればそれを返し、なければ新規作成します。
func (r *ConversationSessionRepository) StartOrResume(
	ctx context.Context,
	userID int64,
) (ConversationSession, error) {
	var session ConversationSession
	err := r.database.QueryRow(
		ctx,
		`SELECT id, started_at, ended_at
		   FROM conversation_sessions
		  WHERE user_id = $1 AND ended_at IS NULL
		  ORDER BY started_at DESC
		  LIMIT 1`,
		userID,
	).Scan(&session.ID, &session.StartedAt, &session.EndedAt)
	if err == nil {
		return session, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ConversationSession{}, fmt.Errorf("find active session: %w", err)
	}

	err = r.database.QueryRow(
		ctx,
		`INSERT INTO conversation_sessions (user_id)
		 VALUES ($1)
		 RETURNING id, started_at, ended_at`,
		userID,
	).Scan(&session.ID, &session.StartedAt, &session.EndedAt)
	if err != nil {
		return ConversationSession{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (r *ConversationSessionRepository) End(
	ctx context.Context,
	userID int64,
	sessionID int64,
) error {
	tag, err := r.database.Exec(
		ctx,
		`UPDATE conversation_sessions
		    SET ended_at = CURRENT_TIMESTAMP,
		        updated_at = CURRENT_TIMESTAMP
		  WHERE id = $1 AND user_id = $2 AND ended_at IS NULL`,
		sessionID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

func (r *ConversationSessionRepository) AddUtterance(
	ctx context.Context,
	userID int64,
	sessionID int64,
	speaker string,
	text string,
	spokenAt time.Time,
) (Utterance, error) {
	var exists bool
	err := r.database.QueryRow(
		ctx,
		`SELECT EXISTS (
		     SELECT 1 FROM conversation_sessions
		      WHERE id = $1 AND user_id = $2
		 )`,
		sessionID,
		userID,
	).Scan(&exists)
	if err != nil {
		return Utterance{}, fmt.Errorf("verify session owner: %w", err)
	}
	if !exists {
		return Utterance{}, ErrSessionNotFound
	}

	utterance := Utterance{Speaker: speaker, Text: text}
	err = r.database.QueryRow(
		ctx,
		`INSERT INTO conversation_utterances
		     (session_id, speaker, text, spoken_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, spoken_at`,
		sessionID,
		speaker,
		text,
		spokenAt,
	).Scan(&utterance.ID, &utterance.SpokenAt)
	if err != nil {
		return Utterance{}, fmt.Errorf("insert utterance: %w", err)
	}
	return utterance, nil
}

func (r *ConversationSessionRepository) List(
	ctx context.Context,
	userID int64,
) ([]ConversationSession, error) {
	rows, err := r.database.Query(
		ctx,
		`SELECT id, started_at, ended_at
		   FROM conversation_sessions
		  WHERE user_id = $1
		  ORDER BY started_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]ConversationSession, 0)
	for rows.Next() {
		var session ConversationSession
		if err := rows.Scan(
			&session.ID,
			&session.StartedAt,
			&session.EndedAt,
		); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *ConversationSessionRepository) ListUtterances(
	ctx context.Context,
	sessionID int64,
) ([]Utterance, error) {
	rows, err := r.database.Query(
		ctx,
		`SELECT id, speaker, text, spoken_at
		   FROM conversation_utterances
		  WHERE session_id = $1
		  ORDER BY spoken_at`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list utterances: %w", err)
	}
	defer rows.Close()

	utterances := make([]Utterance, 0)
	for rows.Next() {
		var utterance Utterance
		if err := rows.Scan(
			&utterance.ID,
			&utterance.Speaker,
			&utterance.Text,
			&utterance.SpokenAt,
		); err != nil {
			return nil, fmt.Errorf("scan utterance: %w", err)
		}
		utterances = append(utterances, utterance)
	}
	return utterances, rows.Err()
}