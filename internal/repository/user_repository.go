package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type UserSettings struct {
	UserID                         int64
	TextSize                       string
	ButtonSize                     string
	Contrast                       string
	SuggestionCount                int
	SpeechRate                     string
	SpeechVolume                   int
	SpeechVoice                    string
	UseHistoryForSuggestions       bool
	UseLocationForSuggestions      bool
	UseProfileForSuggestions       bool
	ShowConfirmationAfterSelection bool
	SaveConversationHistory        bool
	AllowExternalCommunication     bool
}

type UserSettingsPatch struct {
	TextSize                       *string
	ButtonSize                     *string
	Contrast                       *string
	SuggestionCount                *int
	SpeechRate                     *string
	SpeechVolume                   *int
	SpeechVoice                    *string
	UseHistoryForSuggestions       *bool
	UseLocationForSuggestions      *bool
	UseProfileForSuggestions       *bool
	ShowConfirmationAfterSelection *bool
	SaveConversationHistory        *bool
	AllowExternalCommunication     *bool
}

type rowScanner interface {
	Scan(dest ...any) error
}

type UserRepository struct {
	database *pgxpool.Pool
}

func NewUserRepository(database *pgxpool.Pool) *UserRepository {
	return &UserRepository{database: database}
}

func (r *UserRepository) GetSettings(
	ctx context.Context,
	userID int64,
) (UserSettings, error) {
	return scanUserSettings(r.database.QueryRow(ctx, `
		SELECT
			user_id,
			text_size,
			button_size,
			contrast,
			suggestion_count,
			speech_rate,
			speech_volume,
			speech_voice,
			use_history_for_suggestions,
			use_location_for_suggestions,
			use_profile_for_suggestions,
			show_confirmation_after_selection,
			save_conversation_history,
			allow_external_communication
		FROM user_settings
		WHERE user_id = $1
	`, userID))
}

func (r *UserRepository) UpdateSettings(
	ctx context.Context,
	userID int64,
	patch UserSettingsPatch,
) (UserSettings, error) {
	return scanUserSettings(r.database.QueryRow(ctx, `
		UPDATE user_settings
		SET
			text_size = COALESCE($2::text, text_size),
			button_size = COALESCE($3::text, button_size),
			contrast = COALESCE($4::text, contrast),
			suggestion_count = COALESCE($5::integer, suggestion_count),
			speech_rate = COALESCE($6::text, speech_rate),
			speech_volume = COALESCE($7::integer, speech_volume),
			speech_voice = COALESCE($8::text, speech_voice),
			use_history_for_suggestions = COALESCE(
				$9::boolean,
				use_history_for_suggestions
			),
			use_location_for_suggestions = COALESCE(
				$10::boolean,
				use_location_for_suggestions
			),
			use_profile_for_suggestions = COALESCE(
				$11::boolean,
				use_profile_for_suggestions
			),
			show_confirmation_after_selection = COALESCE(
				$12::boolean,
				show_confirmation_after_selection
			),
			save_conversation_history = COALESCE(
				$13::boolean,
				save_conversation_history
			),
			allow_external_communication = COALESCE(
				$14::boolean,
				allow_external_communication
			),
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $1
		RETURNING
			user_id,
			text_size,
			button_size,
			contrast,
			suggestion_count,
			speech_rate,
			speech_volume,
			speech_voice,
			use_history_for_suggestions,
			use_location_for_suggestions,
			use_profile_for_suggestions,
			show_confirmation_after_selection,
			save_conversation_history,
			allow_external_communication
	`,
		userID,
		patch.TextSize,
		patch.ButtonSize,
		patch.Contrast,
		patch.SuggestionCount,
		patch.SpeechRate,
		patch.SpeechVolume,
		patch.SpeechVoice,
		patch.UseHistoryForSuggestions,
		patch.UseLocationForSuggestions,
		patch.UseProfileForSuggestions,
		patch.ShowConfirmationAfterSelection,
		patch.SaveConversationHistory,
		patch.AllowExternalCommunication,
	))
}

func scanUserSettings(row rowScanner) (UserSettings, error) {
	var settings UserSettings
	err := row.Scan(
		&settings.UserID,
		&settings.TextSize,
		&settings.ButtonSize,
		&settings.Contrast,
		&settings.SuggestionCount,
		&settings.SpeechRate,
		&settings.SpeechVolume,
		&settings.SpeechVoice,
		&settings.UseHistoryForSuggestions,
		&settings.UseLocationForSuggestions,
		&settings.UseProfileForSuggestions,
		&settings.ShowConfirmationAfterSelection,
		&settings.SaveConversationHistory,
		&settings.AllowExternalCommunication,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserSettings{}, ErrUserNotFound
	}
	return settings, err
}
