package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/repository"
	"kokotoba-backend/internal/view"
)

type userSettingsRepository interface {
	GetSettings(context.Context, int64) (repository.UserSettings, error)
	UpdateSettings(
		context.Context,
		int64,
		repository.UserSettingsPatch,
	) (repository.UserSettings, error)
}

type UserSettingsController struct {
	repository userSettingsRepository
}

func NewUserSettingsController(repository userSettingsRepository) *UserSettingsController {
	return &UserSettingsController{repository: repository}
}

func (c *UserSettingsController) Show(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	settings, err := c.repository.GetSettings(r.Context(), userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		_ = view.JSON(w, http.StatusNotFound, model.ErrorResponse{
			Message: "user not found",
		})
		return
	}
	if err != nil {
		_ = view.JSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Message: "failed to fetch user settings",
		})
		return
	}

	_ = view.JSON(w, http.StatusOK, model.NewUserSettingsResponse(settings))
}

func (c *UserSettingsController) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseUserID(w, r)
	if !ok {
		return
	}

	request, err := decodeUpdateRequest(w, r)
	if err != nil {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: err.Error(),
		})
		return
	}
	if err := validateUpdateRequest(&request); err != nil {
		_ = view.JSON(w, http.StatusUnprocessableEntity, model.ErrorResponse{
			Message: err.Error(),
		})
		return
	}

	settings, err := c.repository.UpdateSettings(
		r.Context(),
		userID,
		request.ToPatch(),
	)
	if errors.Is(err, repository.ErrUserNotFound) {
		_ = view.JSON(w, http.StatusNotFound, model.ErrorResponse{
			Message: "user not found",
		})
		return
	}
	if err != nil {
		_ = view.JSON(w, http.StatusInternalServerError, model.ErrorResponse{
			Message: "failed to update user settings",
		})
		return
	}

	_ = view.JSON(w, http.StatusOK, model.NewUserSettingsResponse(settings))
}

func parseUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil || userID <= 0 {
		_ = view.JSON(w, http.StatusBadRequest, model.ErrorResponse{
			Message: "userID must be a positive integer",
		})
		return 0, false
	}
	return userID, true
}

func decodeUpdateRequest(
	w http.ResponseWriter,
	r *http.Request,
) (model.UpdateUserSettingsRequest, error) {
	var request model.UpdateUserSettingsRequest
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

func validateUpdateRequest(request *model.UpdateUserSettingsRequest) error {
	if request.TextSize == nil &&
		request.ButtonSize == nil &&
		request.Contrast == nil &&
		request.SuggestionCount == nil &&
		request.SpeechRate == nil &&
		request.SpeechVolume == nil &&
		request.SpeechVoice == nil &&
		request.UseHistoryForSuggestions == nil &&
		request.UseLocationForSuggestions == nil &&
		request.UseProfileForSuggestions == nil &&
		request.ShowConfirmationAfterSelection == nil &&
		request.SaveConversationHistory == nil &&
		request.AllowExternalCommunication == nil {
		return errors.New("at least one setting must be provided")
	}

	for name, value := range map[string]**string{
		"text_size":    &request.TextSize,
		"button_size":  &request.ButtonSize,
		"contrast":     &request.Contrast,
		"speech_rate":  &request.SpeechRate,
		"speech_voice": &request.SpeechVoice,
	} {
		if *value == nil {
			continue
		}
		trimmed := strings.TrimSpace(**value)
		if trimmed == "" {
			return errors.New(name + " must not be empty")
		}
		**value = trimmed
	}

	if request.SuggestionCount != nil && *request.SuggestionCount <= 0 {
		return errors.New("suggestion_count must be greater than zero")
	}
	if request.SpeechVolume != nil &&
		(*request.SpeechVolume < 0 || *request.SpeechVolume > 100) {
		return errors.New("speech_volume must be between 0 and 100")
	}
	return nil
}
