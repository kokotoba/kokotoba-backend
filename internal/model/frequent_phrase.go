package model

import "kokotoba-backend/internal/repository"

type FrequentPhraseResponse struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type FrequentPhraseListResponse struct {
	Phrases []FrequentPhraseResponse `json:"phrases"`
}

type CreateFrequentPhraseRequest struct {
	Text string `json:"text"`
}

type ReorderFrequentPhrasesRequest struct {
	PhraseIDs []int64 `json:"phrase_ids"`
}

func NewFrequentPhraseResponse(
	phrase repository.FrequentPhrase,
) FrequentPhraseResponse {
	return FrequentPhraseResponse{ID: phrase.ID, Text: phrase.Text}
}

func NewFrequentPhraseListResponse(
	phrases []repository.FrequentPhrase,
) FrequentPhraseListResponse {
	response := make([]FrequentPhraseResponse, 0, len(phrases))
	for _, phrase := range phrases {
		response = append(response, NewFrequentPhraseResponse(phrase))
	}
	return FrequentPhraseListResponse{Phrases: response}
}
