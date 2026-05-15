package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportCenterModeDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportCenterModeDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCenterModeDetailLogic {
	return &ReportCenterModeDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportCenterModeDetailLogic) ReportCenterModeDetail(req *types.ReportCenterModeDetailReq) (*types.ReportCenterModeDetailResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	modeKey, err := normalizeReportCenterModeKeyFilter(req.ModeKey)
	if err != nil {
		return nil, err
	}
	limit := normalizeReportCenterSessionsLimit(req.Limit)

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}

	resp := buildReportCenterModeDetail(rows, modeKey, limit)
	return &resp, nil
}

func buildReportCenterModeDetail(rows []reportCenterOverviewRow, modeKey string, limit int64) types.ReportCenterModeDetailResp {
	cards := buildReportCenterModeCards(rows)
	card := findReportCenterModeCard(cards, modeKey)
	reports := filterReportCenterReportsByMode(rows, modeKey)
	total := int64(len(reports))
	hasMore := total > limit
	if total > limit {
		reports = reports[:limit]
	}

	return types.ReportCenterModeDetailResp{
		Card:    card,
		Reports: reports,
		Total:   total,
		Filters: types.ReportCenterModeDetailFilters{
			Mode:    sessionmode.DisplayName(modeKey),
			ModeKey: modeKey,
			Limit:   limit,
			HasMore: hasMore,
		},
		ModeMeta: types.ReportMeta{
			SchemaVersion: "report-center-mode-detail-v1",
			Available:     true,
		},
	}
}

func findReportCenterModeCard(cards []types.ReportCenterModeCard, modeKey string) types.ReportCenterModeCard {
	for _, card := range cards {
		if card.ModeKey == modeKey {
			return card
		}
	}
	return types.ReportCenterModeCard{
		Mode:           sessionmode.DisplayName(modeKey),
		ModeKey:        modeKey,
		AttentionState: "empty",
	}
}

func filterReportCenterReportsByMode(rows []reportCenterOverviewRow, modeKey string) []types.ReportCenterRecentReport {
	reports := make([]types.ReportCenterRecentReport, 0, len(rows))
	for _, row := range rows {
		if sessionmode.NormalizeKey(row.Mode) != modeKey {
			continue
		}
		reports = append(reports, buildReportCenterRecentReport(row))
	}
	return reports
}
