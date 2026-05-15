package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportCenterModesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportCenterModesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportCenterModesLogic {
	return &ReportCenterModesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportCenterModesLogic) ReportCenterModes(_ *types.ReportCenterModesReq) (*types.ReportCenterModesResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	rows, err := fetchReportCenterOverviewRows(l.ctx, l.svcCtx, userID)
	if err != nil {
		return nil, err
	}

	return &types.ReportCenterModesResp{
		Modes: buildReportCenterModeCards(rows),
		ModesMeta: types.ReportMeta{
			SchemaVersion: "report-center-modes-v1",
			Available:     true,
		},
	}, nil
}
