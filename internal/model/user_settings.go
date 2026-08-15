package model

import (
	"strconv"

	"kokotoba-backend/internal/repository"
)

type SettingRow struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SettingToggle struct {
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

type UserSettingsResponse struct {
	UserID         int64           `json:"user_id"`
	DisplayRows    []SettingRow    `json:"display_rows"`
	VoiceRows      []SettingRow    `json:"voice_rows"`
	SupportToggles []SettingToggle `json:"support_toggles"`
	PrivacyToggles []SettingToggle `json:"privacy_toggles"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type UpdateUserSettingsRequest struct {
	TextSize                       *string `json:"text_size"`
	ButtonSize                     *string `json:"button_size"`
	Contrast                       *string `json:"contrast"`
	SuggestionCount                *int    `json:"suggestion_count"`
	SpeechRate                     *string `json:"speech_rate"`
	SpeechVolume                   *int    `json:"speech_volume"`
	SpeechVoice                    *string `json:"speech_voice"`
	UseHistoryForSuggestions       *bool   `json:"use_history_for_suggestions"`
	UseLocationForSuggestions      *bool   `json:"use_location_for_suggestions"`
	UseProfileForSuggestions       *bool   `json:"use_profile_for_suggestions"`
	ShowConfirmationAfterSelection *bool   `json:"show_confirmation_after_selection"`
	SaveConversationHistory        *bool   `json:"save_conversation_history"`
	AllowExternalCommunication     *bool   `json:"allow_external_communication"`
}

func (request UpdateUserSettingsRequest) ToPatch() repository.UserSettingsPatch {
	return repository.UserSettingsPatch{
		TextSize:                       request.TextSize,
		ButtonSize:                     request.ButtonSize,
		Contrast:                       request.Contrast,
		SuggestionCount:                request.SuggestionCount,
		SpeechRate:                     request.SpeechRate,
		SpeechVolume:                   request.SpeechVolume,
		SpeechVoice:                    request.SpeechVoice,
		UseHistoryForSuggestions:       request.UseHistoryForSuggestions,
		UseLocationForSuggestions:      request.UseLocationForSuggestions,
		UseProfileForSuggestions:       request.UseProfileForSuggestions,
		ShowConfirmationAfterSelection: request.ShowConfirmationAfterSelection,
		SaveConversationHistory:        request.SaveConversationHistory,
		AllowExternalCommunication:     request.AllowExternalCommunication,
	}
}

func NewUserSettingsResponse(settings repository.UserSettings) UserSettingsResponse {
	return UserSettingsResponse{
		UserID: settings.UserID,
		DisplayRows: []SettingRow{
			{Label: "文字サイズ", Value: settings.TextSize},
			{Label: "ボタンサイズ", Value: settings.ButtonSize},
			{Label: "コントラスト", Value: settings.Contrast},
			{Label: "候補表示数", Value: strconv.Itoa(settings.SuggestionCount) + "件"},
		},
		VoiceRows: []SettingRow{
			{Label: "読み上げ速度", Value: settings.SpeechRate},
			{Label: "読み上げ音量", Value: strconv.Itoa(settings.SpeechVolume) + "%"},
			{Label: "読み上げ音声", Value: settings.SpeechVoice},
		},
		SupportToggles: []SettingToggle{
			{Label: "履歴を候補に利用", Enabled: settings.UseHistoryForSuggestions},
			{Label: "位置情報を候補に利用", Enabled: settings.UseLocationForSuggestions},
			{Label: "登録情報を候補に利用", Enabled: settings.UseProfileForSuggestions},
			{Label: "選択後に確認画面を表示", Enabled: settings.ShowConfirmationAfterSelection},
		},
		PrivacyToggles: []SettingToggle{
			{Label: "会話履歴を保存", Enabled: settings.SaveConversationHistory},
			{Label: "外部通信を利用", Enabled: settings.AllowExternalCommunication},
		},
	}
}
