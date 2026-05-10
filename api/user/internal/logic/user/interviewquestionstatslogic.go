package user

import (
	"context"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InterviewQuestionStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInterviewQuestionStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InterviewQuestionStatsLogic {
	return &InterviewQuestionStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InterviewQuestionStatsLogic) InterviewQuestionStats(_ *types.InterviewQuestionStatsReq) (*types.InterviewQuestionStatsResp, error) {
	if _, err := currentUserID(l.ctx); err != nil {
		return nil, err
	}

	stats, err := l.svcCtx.InterviewQuestionsModel.Stats(l.ctx)
	if err != nil {
		return nil, err
	}

	directions := make([]types.InterviewQuestionDirectionStat, 0, len(stats.Directions))
	for _, item := range stats.Directions {
		directions = append(directions, types.InterviewQuestionDirectionStat{
			DirectionKey: item.DirectionKey,
			Count:        item.Count,
		})
	}
	difficulties := make([]types.InterviewQuestionDifficultyStat, 0, len(stats.Difficulties))
	for _, item := range stats.Difficulties {
		difficulties = append(difficulties, types.InterviewQuestionDifficultyStat{
			DifficultyLevel: item.DifficultyLevel,
			Count:           item.Count,
		})
	}

	resp := &types.InterviewQuestionStatsResp{
		Total:        stats.Total,
		Directions:   directions,
		Difficulties: difficulties,
		StatsMeta: types.ReportMeta{
			SchemaVersion: interviewQuestionSchemaVersion,
			Available:     stats.Total > 0,
		},
	}
	if stats.UpdatedAt.Valid {
		resp.UpdatedAt = stats.UpdatedAt.Time.Format(timeLayout)
	}
	return resp, nil
}
