package user

import (
	"encoding/json"

	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/api/user/model"
	"GoZero-AI/internal/sessionmode"
)

func buildSessionItem(session model.ChatSession) types.SessionItem {
	modeKey := normalizeSessionMode(session.Mode)

	item := types.SessionItem{
		SessionId:    session.SessionId,
		Title:        session.Title,
		Mode:         sessionmode.DisplayName(modeKey),
		ModeKey:      modeKey,
		MessageCount: session.MessageCount,
		IsActive:     session.IsActive,
		CreatedAt:    session.CreatedAt.Format(timeLayout),
		UpdatedAt:    session.UpdatedAt.Format(timeLayout),
	}
	if session.LastMessageAt.Valid {
		item.LastMessageAt = session.LastMessageAt.Time.Format(timeLayout)
	}
	if session.CompletedAt.Valid {
		item.CompletedAt = session.CompletedAt.Time.Format(timeLayout)
	}

	return item
}

func buildSessionConfigSnapshot(session model.ChatSession) types.SessionConfigSnapshot {
	focusAreas := decodeSessionFocusAreas(session.FocusAreas)
	if len(focusAreas) == 0 {
		focusAreas = defaultFocusAreasForDirection(defaultStringValue(session.DirectionKey, "go_backend"))
	}

	config := types.SessionConfigSnapshot{
		DirectionKey:          defaultStringValue(session.DirectionKey, "go_backend"),
		DirectionLabel:        defaultStringValue(session.DirectionLabel, "Go 后端"),
		DifficultyLevel:       defaultInt64Value(session.DifficultyLevel, 3),
		DifficultyLabel:       defaultStringValue(session.DifficultyLabel, "中级"),
		InterviewerStyle:      defaultStringValue(session.InterviewerStyle, "senior"),
		InterviewerStyleLabel: defaultStringValue(session.InterviewerStyleLabel, "资深技术官"),
		FocusAreas:            focusAreas,
		FollowUpDepth:         defaultStringValue(session.FollowUpDepth, "N+3"),
		EstimatedMinutes:      defaultInt64Value(session.EstimatedMinutes, 30),
		ProgressPercent:       session.ProgressPercent,
		DurationSeconds:       session.DurationSeconds,
	}
	if session.StartedAt.Valid {
		config.StartedAt = session.StartedAt.Time.Format(timeLayout)
	}
	if session.CompletedAt.Valid {
		config.CompletedAt = session.CompletedAt.Time.Format(timeLayout)
	}
	return config
}

func decodeSessionFocusAreas(raw []byte) []types.FocusAreaSelection {
	if len(raw) == 0 {
		return nil
	}

	var focusAreas []types.FocusAreaSelection
	if err := json.Unmarshal(raw, &focusAreas); err != nil {
		return nil
	}
	return focusAreas
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

const defaultSessionMode = sessionmode.DefaultKey

func normalizeSessionMode(value string) string {
	return sessionmode.NormalizeKey(value)
}

func defaultStringValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultInt64Value(value, fallback int64) int64 {
	if value == 0 {
		return fallback
	}
	return value
}
