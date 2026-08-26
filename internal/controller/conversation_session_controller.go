package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/repository"
	"kokotoba-backend/internal/view"
)

type conversationSessionRepository interface {
	StartOrResume(context.Context, int64) (repository.ConversationSession, error)
	End(context.Context, int64, int64) error
	AddUtterance(
		context.Context,
		int64,
		int64,
		string,
		string,
		time.Time,
	) (repository.Utterance, error)
	List(context.Context, int64) ([]repository.ConversationSession, error)
	ListUtterances(context.Context, int64) ([]repository.Utterance, error)
}

type ConversationSessionController struct {
	repository conversationSessionRepository
}

func NewConversationSessionController(
	repository conversationSessionRepository,
) *ConversationSessionController {
	return &ConversationSessionController{repository: repository}
}

// Create は進行中セッションがあれば再開し、なければ新規作成します。
func (c *ConversationSessionController) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	session, err := c.repository.StartOrResume(r.Context(), userID)
	if handleSessionRepositoryError(w, err) {
		return
	}

	utterances, err := c.repository.ListUtterances(r.Context(), session.ID)
	if handleSessionRepositoryError(w, err) {
		return
	}

	_ = view.JSON(
		w,
		http.StatusOK,
		model.NewConversationSessionResponse(session, utterances),
	)
}

func (c *ConversationSessionController) Index(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	sessions, err := c.repository.List(r.Context(), userID)
	if handleSessionRepositoryError(w, err) {
		return
	}

	responses := make([]model.ConversationSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		utterances, err := c.repository.ListUtterances(r.Context(), session.ID)
		if handleSessionRepositoryError(w, err) {
			return
		}
		responses = append(
			responses,
			model.NewConversationSessionResponse(session, utterances),
		)
	}

	_ = view.JSON(
		w,
		http.StatusOK,
		model.NewConversationSessionListResponse(responses),
	)
}

func (c *ConversationSessionController) End(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	sessionID, ok := sessionIDFromPath(w, r)
	if !ok {
		return
	}

	if handleSessionRepositoryError(
		w,
		c.repository.End(r.Context(), userID, sessionID),
	) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ConversationSessionController) CreateUtterance(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := authenticatedUserID(w, r)
	if !ok {
		return
	}

	sessionID, ok := sessionIDFromPath(w, r)
	if !ok {
		return
	}

	request, err := decodeCreateUtteranceRequest(w, r)
	if err != nil {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	text := strings.TrimSpace(request.Text)
	if text == "" {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: "text must not be empty",
		})
		return
	}
	if utf8.RuneCountInString(text) > 2000 {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: "text must be 2000 characters or fewer",
		})
		return
	}
	if request.Speaker != "partner" && request.Speaker != "user" {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: "speaker must be either partner or user",
		})
		return
	}

	spokenAt := time.Now()
	if request.SpokenAt != nil {
		spokenAt = *request.SpokenAt
	}

	utterance, err := c.repository.AddUtterance(
		r.Context(),
		userID,
		sessionID,
		request.Speaker,
		text,
		spokenAt,
	)
	if handleSessionRepositoryError(w, err) {
		return
	}

	_ = view.JSON(
		w,
		http.StatusCreated,
		model.NewUtteranceResponse(utterance),
	)
}

func decodeCreateUtteranceRequest(
	w http.ResponseWriter,
	r *http.Request,
) (model.CreateUtteranceRequest, error) {
	var request model.CreateUtteranceRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, errors.New("request body must be valid JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return request, errors.New("request body must contain one JSON object")
	}
	return request, nil
}

func sessionIDFromPath(w http.ResponseWriter, r *http.Request) (int64, bool) {
	sessionID, err := strconv.ParseInt(r.PathValue("sessionID"), 10, 64)
	if err != nil || sessionID <= 0 {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: "sessionID must be a positive integer",
		})
		return 0, false
	}
	return sessionID, true
}

func handleSessionRepositoryError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	status := http.StatusInternalServerError
	message := "failed to process conversation session"
	switch {
	case errors.Is(err, repository.ErrSessionNotFound):
		status = http.StatusNotFound
		message = "conversation session not found"
	case errors.Is(err, repository.ErrUserNotFound):
		status = http.StatusNotFound
		message = "user not found"
	}
	_ = view.JSON(w, status, model.ErrorResponse{Message: message})
	return true
}