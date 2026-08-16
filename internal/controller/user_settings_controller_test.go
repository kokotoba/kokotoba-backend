package controller

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kokotoba-backend/internal/authn"
	"kokotoba-backend/internal/repository"
)

type fakeUserSettingsRepository struct {
	settings      repository.UserSettings
	err           error
	receivedUser  int64
	receivedPatch repository.UserSettingsPatch
}

func (r fakeUserSettingsRepository) GetSettings(
	context.Context,
	int64,
) (repository.UserSettings, error) {
	return r.settings, r.err
}

func (r *fakeUserSettingsRepository) UpdateSettings(
	_ context.Context,
	userID int64,
	patch repository.UserSettingsPatch,
) (repository.UserSettings, error) {
	r.receivedUser = userID
	r.receivedPatch = patch
	return r.settings, r.err
}

func TestUserSettingsControllerShow(t *testing.T) {
	controller := NewUserSettingsController(&fakeUserSettingsRepository{
		settings: repository.UserSettings{
			UserID:                   1,
			TextSize:                 "大きい",
			ButtonSize:               "大きい",
			Contrast:                 "標準",
			SuggestionCount:          3,
			SpeechRate:               "標準",
			SpeechVolume:             80,
			SpeechVoice:              "日本語 1",
			UseHistoryForSuggestions: true,
		},
	})
	request := authenticatedRequest(http.MethodGet, "/api/v1/me/settings", nil, 1)
	response := httptest.NewRecorder()

	controller.Show(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	for _, expected := range []string{
		`"user_id":1`,
		`"label":"文字サイズ"`,
		`"value":"3件"`,
		`"enabled":true`,
	} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Errorf("response body does not contain %q: %s", expected, response.Body.String())
		}
	}
}

func TestUserSettingsControllerRequiresAuthenticatedUser(t *testing.T) {
	controller := NewUserSettingsController(&fakeUserSettingsRepository{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil)
	response := httptest.NewRecorder()

	controller.Show(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestUserSettingsControllerReturnsNotFound(t *testing.T) {
	controller := NewUserSettingsController(&fakeUserSettingsRepository{
		err: repository.ErrUserNotFound,
	})
	request := authenticatedRequest(http.MethodGet, "/api/v1/me/settings", nil, 99)
	response := httptest.NewRecorder()

	controller.Show(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestUserSettingsControllerReturnsInternalServerError(t *testing.T) {
	controller := NewUserSettingsController(&fakeUserSettingsRepository{
		err: errors.New("database unavailable"),
	})
	request := authenticatedRequest(http.MethodGet, "/api/v1/me/settings", nil, 1)
	response := httptest.NewRecorder()

	controller.Show(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			response.Code,
			http.StatusInternalServerError,
		)
	}
}

func TestUserSettingsControllerUpdate(t *testing.T) {
	storage := &fakeUserSettingsRepository{
		settings: repository.UserSettings{
			UserID:                   1,
			TextSize:                 "標準",
			ButtonSize:               "大きい",
			Contrast:                 "標準",
			SuggestionCount:          5,
			SpeechRate:               "標準",
			SpeechVolume:             80,
			SpeechVoice:              "日本語 1",
			UseHistoryForSuggestions: false,
		},
	}
	controller := NewUserSettingsController(storage)
	request := authenticatedRequest(
		http.MethodPatch,
		"/api/v1/me/settings",
		strings.NewReader(`{
			"text_size":" 標準 ",
			"suggestion_count":5,
			"use_history_for_suggestions":false
		}`),
		1,
	)
	response := httptest.NewRecorder()

	controller.Update(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if storage.receivedUser != 1 {
		t.Errorf("userID = %d, want 1", storage.receivedUser)
	}
	if storage.receivedPatch.TextSize == nil ||
		*storage.receivedPatch.TextSize != "標準" {
		t.Errorf("text size was not trimmed: %#v", storage.receivedPatch.TextSize)
	}
	if storage.receivedPatch.UseHistoryForSuggestions == nil ||
		*storage.receivedPatch.UseHistoryForSuggestions {
		t.Error("explicit false value was not preserved")
	}
	if !strings.Contains(response.Body.String(), `"value":"5件"`) {
		t.Errorf("response does not contain updated value: %s", response.Body.String())
	}
}

func TestUserSettingsControllerUpdateRejectsInvalidBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty patch", body: `{}`},
		{name: "invalid volume", body: `{"speech_volume":101}`},
		{name: "blank text", body: `{"text_size":" "}`},
		{name: "unknown field", body: `{"unknown":true}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := NewUserSettingsController(&fakeUserSettingsRepository{})
			request := authenticatedRequest(
				http.MethodPatch,
				"/api/v1/me/settings",
				strings.NewReader(test.body),
				1,
			)
			response := httptest.NewRecorder()

			controller.Update(response, request)

			if response.Code != http.StatusBadRequest &&
				response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 400 or 422", response.Code)
			}
		})
	}
}

func authenticatedRequest(
	method string,
	target string,
	body io.Reader,
	userID int64,
) *http.Request {
	request := httptest.NewRequest(method, target, body)
	return request.WithContext(authn.ContextWithUser(request.Context(), authn.User{ID: userID}))
}
