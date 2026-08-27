package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kokotoba-backend/internal/repository"
)

type fakeConversationSessionRepository struct {
	session       repository.ConversationSession
	sessions      []repository.ConversationSession
	utterances    []repository.Utterance
	receivedUser  int64
	receivedID    int64
	receivedRole  string
	receivedText  string
	receivedAt    time.Time
	endedSession  int64
	repositoryErr error
}

func (r *fakeConversationSessionRepository) StartOrResume(
	_ context.Context,
	userID int64,
) (repository.ConversationSession, error) {
	r.receivedUser = userID
	return r.session, r.repositoryErr
}

func (r *fakeConversationSessionRepository) End(
	_ context.Context,
	_ int64,
	sessionID int64,
) error {
	r.endedSession = sessionID
	return r.repositoryErr
}

func (r *fakeConversationSessionRepository) AddUtterance(
	_ context.Context,
	userID int64,
	sessionID int64,
	speaker string,
	text string,
	spokenAt time.Time,
) (repository.Utterance, error) {
	r.receivedUser = userID
	r.receivedID = sessionID
	r.receivedRole = speaker
	r.receivedText = text
	r.receivedAt = spokenAt
	return repository.Utterance{
		ID:       9,
		Speaker:  speaker,
		Text:     text,
		SpokenAt: spokenAt,
	}, r.repositoryErr
}

func (r *fakeConversationSessionRepository) List(
	context.Context,
	int64,
) ([]repository.ConversationSession, error) {
	return r.sessions, r.repositoryErr
}

func (r *fakeConversationSessionRepository) ListUtterances(
	context.Context,
	int64,
) ([]repository.Utterance, error) {
	return r.utterances, r.repositoryErr
}

func TestConversationSessionControllerStartsOrResumesSession(t *testing.T) {
	startedAt := time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)
	storage := &fakeConversationSessionRepository{
		session: repository.ConversationSession{ID: 7, StartedAt: startedAt},
		utterances: []repository.Utterance{
			{
				ID:       1,
				Speaker:  "partner",
				Text:     "体調はいかがですか？",
				SpokenAt: startedAt,
			},
		},
	}
	controller := NewConversationSessionController(storage)
	request := authenticatedRequest(
		http.MethodPost,
		"/api/v1/me/sessions",
		nil,
		1,
	)
	response := httptest.NewRecorder()

	controller.Create(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if storage.receivedUser != 1 {
		t.Fatalf("user ID = %d, want 1", storage.receivedUser)
	}
	for _, expected := range []string{
		`"id":7`,
		`"active":true`,
		`"speaker":"partner"`,
		`"text":"体調はいかがですか？"`,
	} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Errorf("response does not contain %s: %s", expected, response.Body.String())
		}
	}
}

func TestConversationSessionControllerSavesUserUtterance(t *testing.T) {
	storage := &fakeConversationSessionRepository{}
	controller := NewConversationSessionController(storage)
	request := authenticatedRequest(
		http.MethodPost,
		"/api/v1/me/sessions/7/utterances",
		strings.NewReader(`{"speaker":"user","text":"  お願いします  "}`),
		1,
	)
	request.SetPathValue("sessionID", "7")
	response := httptest.NewRecorder()

	controller.CreateUtterance(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if storage.receivedID != 7 || storage.receivedRole != "user" {
		t.Fatalf("session = %d, speaker = %q", storage.receivedID, storage.receivedRole)
	}
	if storage.receivedText != "お願いします" {
		t.Fatalf("text = %q, want trimmed text", storage.receivedText)
	}
}

func TestConversationSessionControllerRejectsUnknownSpeaker(t *testing.T) {
	storage := &fakeConversationSessionRepository{}
	controller := NewConversationSessionController(storage)
	request := authenticatedRequest(
		http.MethodPost,
		"/api/v1/me/sessions/7/utterances",
		strings.NewReader(`{"speaker":"assistant","text":"test"}`),
		1,
	)
	request.SetPathValue("sessionID", "7")
	response := httptest.NewRecorder()

	controller.CreateUtterance(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	if storage.receivedText != "" {
		t.Fatal("repository must not be called for an invalid speaker")
	}
}
