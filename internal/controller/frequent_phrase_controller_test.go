package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kokotoba-backend/internal/repository"
)

type fakeFrequentPhraseRepository struct {
	phrases      []repository.FrequentPhrase
	err          error
	createdText  string
	deletedID    int64
	reorderedIDs []int64
}

func (r *fakeFrequentPhraseRepository) List(
	context.Context,
	int64,
) ([]repository.FrequentPhrase, error) {
	return r.phrases, r.err
}

func (r *fakeFrequentPhraseRepository) Create(
	_ context.Context,
	_ int64,
	text string,
) (repository.FrequentPhrase, error) {
	r.createdText = text
	if r.err != nil {
		return repository.FrequentPhrase{}, r.err
	}
	return repository.FrequentPhrase{ID: 10, Text: text}, nil
}

func (r *fakeFrequentPhraseRepository) Delete(
	_ context.Context,
	_ int64,
	phraseID int64,
) error {
	r.deletedID = phraseID
	return r.err
}

func (r *fakeFrequentPhraseRepository) Reorder(
	_ context.Context,
	_ int64,
	phraseIDs []int64,
) error {
	r.reorderedIDs = phraseIDs
	return r.err
}

func TestFrequentPhraseControllerIndexReturnsEmptyList(t *testing.T) {
	controller := NewFrequentPhraseController(&fakeFrequentPhraseRepository{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/1/phrases", nil)
	request.SetPathValue("userID", "1")
	response := httptest.NewRecorder()

	controller.Index(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(response.Body.String()); body != `{"phrases":[]}` {
		t.Fatalf("body = %s, want empty phrases", body)
	}
}

func TestFrequentPhraseControllerCreateTrimsText(t *testing.T) {
	storage := &fakeFrequentPhraseRepository{}
	controller := NewFrequentPhraseController(storage)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/users/1/phrases",
		strings.NewReader(`{"text":"  もう一度お願いします  "}`),
	)
	request.SetPathValue("userID", "1")
	response := httptest.NewRecorder()

	controller.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if storage.createdText != "もう一度お願いします" {
		t.Fatalf("created text = %q", storage.createdText)
	}
}

func TestFrequentPhraseControllerCreateRejectsEmptyText(t *testing.T) {
	controller := NewFrequentPhraseController(&fakeFrequentPhraseRepository{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/users/1/phrases",
		strings.NewReader(`{"text":" "}`),
	)
	request.SetPathValue("userID", "1")
	response := httptest.NewRecorder()

	controller.Create(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusUnprocessableEntity,
		)
	}
}

func TestFrequentPhraseControllerDelete(t *testing.T) {
	storage := &fakeFrequentPhraseRepository{}
	controller := NewFrequentPhraseController(storage)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/users/1/phrases/12",
		nil,
	)
	request.SetPathValue("userID", "1")
	request.SetPathValue("phraseID", "12")
	response := httptest.NewRecorder()

	controller.Delete(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if storage.deletedID != 12 {
		t.Fatalf("deleted ID = %d, want 12", storage.deletedID)
	}
}

func TestFrequentPhraseControllerReorder(t *testing.T) {
	storage := &fakeFrequentPhraseRepository{}
	controller := NewFrequentPhraseController(storage)
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/users/1/phrases/order",
		strings.NewReader(`{"phrase_ids":[3,1,2]}`),
	)
	request.SetPathValue("userID", "1")
	response := httptest.NewRecorder()

	controller.Reorder(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := storage.reorderedIDs; len(got) != 3 || got[0] != 3 || got[1] != 1 || got[2] != 2 {
		t.Fatalf("reordered IDs = %v, want [3 1 2]", got)
	}
}

func TestFrequentPhraseControllerReorderRejectsDuplicates(t *testing.T) {
	storage := &fakeFrequentPhraseRepository{}
	controller := NewFrequentPhraseController(storage)
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/users/1/phrases/order",
		strings.NewReader(`{"phrase_ids":[1,1]}`),
	)
	request.SetPathValue("userID", "1")
	response := httptest.NewRecorder()

	controller.Reorder(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusUnprocessableEntity,
		)
	}
	if storage.reorderedIDs != nil {
		t.Fatalf("repository was called with %v", storage.reorderedIDs)
	}
}
