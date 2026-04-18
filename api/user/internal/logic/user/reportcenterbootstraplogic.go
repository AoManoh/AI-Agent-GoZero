package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportCenterBootstrapLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportCenterBootstrapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCenterBootstrapLogic {
	return &ReportCenterBootstrapLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportCenterBootstrapLogic) ReportCenterBootstrap(req *types.ReportCenterBootstrapReq) (*types.ReportCenterBootstrapResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	modeKey, err := resolveReportCenterModeFilter(req.Mode, req.ModeKey)
	if err != nil {
		return nil, err
	}
	if modeKey == "" {
		modeKey = sessionmode.DefaultKey
	}
	limit := normalizeReportCenterSessionsLimit(req.Limit)

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}

	overviewTotals, overviewModes, recentReports := buildReportCenterOverview(rows)
	modeCards := buildReportCenterModeCards(rows)
	modeDetail := buildReportCenterModeDetail(rows, modeKey, limit)

	return &types.ReportCenterBootstrapResp{
		Overview: types.ReportCenterOverviewResp{
			Totals:        overviewTotals,
			Modes:         overviewModes,
			RecentReports: recentReports,
			OverviewMeta: types.ReportMeta{
				SchemaVersion: "report-center-overview-v1",
				Available:     true,
			},
		},
		Modes:      modeCards,
		ModeDetail: modeDetail,
		BootstrapMeta: types.ReportMeta{
			SchemaVersion: "report-center-bootstrap-v1",
			Available:     true,
		},
	}, nil
}
