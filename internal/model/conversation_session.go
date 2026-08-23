package model

import (
	"time"

	"kokotoba-backend/internal/repository"
)

type UtteranceResponse struct {
	ID       int64     `json:"id"`
	Speaker  string    `json:"speaker"`
	Text     string    `json:"text"`
	SpokenAt time.Time `json:"spoken_at"`
}

type ConversationSessionResponse struct {
	ID         int64               `json:"id"`
	StartedAt  time.Time           `json:"started_at"`
	EndedAt    *time.Time          `json:"ended_at"`
	Active     bool                `json:"active"`
	Utterances []UtteranceResponse `json:"utterances"`
}

type ConversationSessionListResponse struct {
	Sessions []ConversationSessionResponse `json:"sessions"`
}

type CreateUtteranceRequest struct {
	Speaker  string     `json:"speaker"`
	Text     string     `json:"text"`
	SpokenAt *time.Time `json:"spoken_at"`
}

func NewUtteranceResponse(
	utterance repository.Utterance,
) UtteranceResponse {
	return UtteranceResponse{
		ID:       utterance.ID,
		Speaker:  utterance.Speaker,
		Text:     utterance.Text,
		SpokenAt: utterance.SpokenAt,
	}
}

func NewConversationSessionResponse(
	session repository.ConversationSession,
	utterances []repository.Utterance,
) ConversationSessionResponse {
	response := make([]UtteranceResponse, 0, len(utterances))
	for _, utterance := range utterances {
		response = append(response, NewUtteranceResponse(utterance))
	}
	return ConversationSessionResponse{
		ID:         session.ID,
		StartedAt:  session.StartedAt,
		EndedAt:    session.EndedAt,
		Active:     session.EndedAt == nil,
		Utterances: response,
	}
}

func NewConversationSessionListResponse(
	sessions []ConversationSessionResponse,
) ConversationSessionListResponse {
	return ConversationSessionListResponse{Sessions: sessions}
}