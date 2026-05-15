package user

import (
	"context"
	"fmt"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultReportCenterSessionsLimit = 20
	maxReportCenterSessionsLimit     = 100
)

type ReportCenterSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportCenterSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCenterSessionsLogic {
	return &ReportCenterSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportCenterSessionsLogic) ReportCenterSessions(req *types.ReportCenterSessionsReq) (*types.ReportCenterSessionsResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	modeKey, err := resolveReportCenterModeFilter(req.Mode, req.ModeKey)
	if err != nil {
		return nil, err
	}
	status, err := normalizeReportCenterStatusFilter(req.Status)
	if err != nil {
		return nil, err
	}
	limit := normalizeReportCenterSessionsLimit(req.Limit)

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}

	filtered := make([]types.ReportCenterRecentReport, 0, len(rows))
	for _, row := range rows {
		rowModeKey := sessionmode.NormalizeKey(row.Mode)
		rowStatus := normalizeOverviewStatus(row.Status)

		if modeKey != "" && rowModeKey != modeKey {
			continue
		}
		if status != "" && rowStatus != status {
			continue
		}

		filtered = append(filtered, buildReportCenterRecentReport(row))
	}

	total := int64(len(filtered))
	hasMore := total > limit
	if total > limit {
		filtered = filtered[:limit]
	}

	return &types.ReportCenterSessionsResp{
		Reports: filtered,
		Total:   total,
		Filters: types.ReportCenterSessionsFilters{
			Mode:    displayFilterMode(modeKey),
			ModeKey: modeKey,
			Status:  status,
			Limit:   limit,
			HasMore: hasMore,
		},
		SessionsMeta: types.ReportMeta{
			SchemaVersion: "report-center-sessions-v1",
			Available:     true,
		},
	}, nil
}

func resolveReportCenterModeFilter(mode, modeKey string) (string, error) {
	if strings.TrimSpace(mode) != "" {
		return normalizeReportCenterModeAlias(mode)
	}
	return normalizeReportCenterModeKeyFilter(modeKey)
}

func normalizeReportCenterModeAlias(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	normalized := sessionmode.NormalizeKey(trimmed)
	if normalized == sessionmode.DefaultKey {
		lower := strings.ToLower(trimmed)
		if lower != "interview" && lower != "interview studio" {
			return "", fmt.Errorf("不支持的 modeKey: %s", trimmed)
		}
	}
	return normalized, nil
}

func normalizeReportCenterModeKeyFilter(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "":
		return "", nil
	case sessionmode.KeyInterview, sessionmode.KeyResearch, sessionmode.KeyMemory, sessionmode.KeyCoach:
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("不支持的 modeKey: %s", value)
	}
}

func normalizeReportCenterStatusFilter(value string) (string, error) {
	switch strings.TrimSpace(value) {
	case "":
		return "", nil
	case "ready", "draft", "insufficient_data":
		return strings.TrimSpace(value), nil
	default:
		return "", fmt.Errorf("不支持的 status: %s", value)
	}
}

func normalizeReportCenterSessionsLimit(value int64) int64 {
	if value <= 0 {
		return defaultReportCenterSessionsLimit
	}
	if value > maxReportCenterSessionsLimit {
		return maxReportCenterSessionsLimit
	}
	return value
}

func displayFilterMode(modeKey string) string {
	if modeKey == "" {
		return ""
	}
	return sessionmode.DisplayName(modeKey)
}
