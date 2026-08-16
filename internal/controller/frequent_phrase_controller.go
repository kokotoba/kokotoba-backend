package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/repository"
	"kokotoba-backend/internal/view"
)

type frequentPhraseRepository interface {
	List(context.Context, int64) ([]repository.FrequentPhrase, error)
	Create(context.Context, int64, string) (repository.FrequentPhrase, error)
	Delete(context.Context, int64, int64) error
	Reorder(context.Context, int64, []int64) error
}

type FrequentPhraseController struct {
	repository frequentPhraseRepository
}

func NewFrequentPhraseController(
	repository frequentPhraseRepository,
) *FrequentPhraseController {
	return &FrequentPhraseController{repository: repository}
}

func (c *FrequentPhraseController) Index(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}
	phrases, err := c.repository.List(r.Context(), userID)
	if handlePhraseRepositoryError(w, err) {
		return
	}
	_ = view.JSON(w, http.StatusOK, model.NewFrequentPhraseListResponse(phrases))
}

func (c *FrequentPhraseController) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	request, err := decodeCreatePhraseRequest(w, r)
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
	if utf8.RuneCountInString(text) > 500 {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: "text must be 500 characters or fewer",
		})
		return
	}

	phrase, err := c.repository.Create(r.Context(), userID, text)
	if handlePhraseRepositoryError(w, err) {
		return
	}
	_ = view.JSON(w, http.StatusCreated, model.NewFrequentPhraseResponse(phrase))
}

func (c *FrequentPhraseController) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}
	phraseID, err := strconv.ParseInt(r.PathValue("phraseID"), 10, 64)
	if err != nil || phraseID <= 0 {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: "phraseID must be a positive integer",
		})
		return
	}
	if handlePhraseRepositoryError(
		w,
		c.repository.Delete(r.Context(), userID, phraseID),
	) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *FrequentPhraseController) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	request, err := decodeReorderPhrasesRequest(w, r)
	if err != nil {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	if request.PhraseIDs == nil {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: "phrase_ids is required",
		})
		return
	}
	seen := make(map[int64]struct{}, len(request.PhraseIDs))
	for _, phraseID := range request.PhraseIDs {
		if phraseID <= 0 {
			_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
				Message: "phrase_ids must contain positive integers",
			})
			return
		}
		if _, duplicate := seen[phraseID]; duplicate {
			_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
				Message: "phrase_ids must not contain duplicates",
			})
			return
		}
		seen[phraseID] = struct{}{}
	}

	if handlePhraseRepositoryError(
		w,
		c.repository.Reorder(r.Context(), userID, request.PhraseIDs),
	) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeCreatePhraseRequest(
	w http.ResponseWriter,
	r *http.Request,
) (model.CreateFrequentPhraseRequest, error) {
	var request model.CreateFrequentPhraseRequest
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

func decodeReorderPhrasesRequest(
	w http.ResponseWriter,
	r *http.Request,
) (model.ReorderFrequentPhrasesRequest, error) {
	var request model.ReorderFrequentPhrasesRequest
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

func handlePhraseRepositoryError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	status := http.StatusInternalServerError
	message := "failed to process frequent phrase"
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		status = http.StatusNotFound
		message = "user not found"
	case errors.Is(err, repository.ErrPhraseNotFound):
		status = http.StatusNotFound
		message = "frequent phrase not found"
	case errors.Is(err, repository.ErrPhraseDuplicate):
		status = http.StatusConflict
		message = "frequent phrase already exists"
	case errors.Is(err, repository.ErrPhraseOrderInvalid):
		status = http.StatusUnprocessableEntity
		message = "phrase_ids must contain every frequent phrase exactly once"
	}
	_ = view.JSON(w, status, model.ErrorResponse{Message: message})
	return true
}
