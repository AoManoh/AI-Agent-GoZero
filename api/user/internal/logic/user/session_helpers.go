package user

import (
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

	return item
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

const defaultSessionMode = sessionmode.DefaultKey

func normalizeSessionMode(value string) string {
	return sessionmode.NormalizeKey(value)
}
