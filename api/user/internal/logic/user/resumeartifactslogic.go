package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/statuserr"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ResumeArtifactsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type ResumeArtifactDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type resumeArtifactRow struct {
	ArtifactId       string    `db:"artifact_id"`
	Title            string    `db:"title"`
	Version          int64     `db:"version"`
	Filename         string    `db:"filename"`
	Status           string    `db:"status"`
	ChunkCount       int64     `db:"chunk_count"`
	BoundSessionName string    `db:"bound_session_name"`
	UploadedAt       time.Time `db:"uploaded_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

type resumeArtifactChunkRow struct {
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func NewResumeArtifactsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeArtifactsLogic {
	return &ResumeArtifactsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewResumeArtifactDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeArtifactDetailLogic {
	return &ResumeArtifactDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResumeArtifactsLogic) ResumeArtifacts(_ *types.ResumeArtifactsReq) (*types.ResumeArtifactsResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	artifacts, err := loadResumeArtifactItems(l.ctx, l.svcCtx.DB, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("简历资料暂不可用，请稍后重试")
	}
	return &types.ResumeArtifactsResp{
		Artifacts: artifacts,
		Total:     int64(len(artifacts)),
	}, nil
}

func (l *ResumeArtifactDetailLogic) ResumeArtifactDetail(req *types.ResumeArtifactDetailReq) (*types.ResumeArtifactDetailResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	artifact, err := loadResumeArtifactItem(l.ctx, l.svcCtx.DB, userID, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
			return nil, statuserr.NotFound("简历资料不存在或已删除")
		}
		return nil, statuserr.ServiceUnavailable("简历资料暂不可用，请稍后重试")
	}

	var rows []resumeArtifactChunkRow
	err = l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, `select content, created_at
from "public"."vector_store"
where user_id = $1 and chat_id = $2 and doc_type = 'resume'
order by created_at asc, id asc`, userID, req.Id)
	if err != nil && err != sqlx.ErrNotFound && err != sql.ErrNoRows {
		return nil, statuserr.ServiceUnavailable("简历分块暂不可用，请稍后重试")
	}

	chunks := make([]types.ResumeArtifactChunk, 0, len(rows))
	for idx, row := range rows {
		chunks = append(chunks, types.ResumeArtifactChunk{
			Index:   int64(idx + 1),
			Content: truncateEvaluationContent(row.Content, 800),
		})
	}

	return &types.ResumeArtifactDetailResp{
		Artifact: artifact,
		Chunks:   chunks,
		Meta: types.ReportMeta{
			SchemaVersion: "resume-artifact-v1",
			Available:     true,
		},
	}, nil
}

func loadResumeArtifactItems(ctx context.Context, db sqlx.SqlConn, userID int64) ([]types.ResumeArtifactItem, error) {
	documentItems, err := loadResumeDocumentItems(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	legacyItems, err := loadLegacyResumeArtifactItems(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(documentItems))
	items := make([]types.ResumeArtifactItem, 0, len(documentItems)+len(legacyItems))
	for _, item := range documentItems {
		seen[item.ArtifactId] = struct{}{}
		items = append(items, item)
	}
	for _, item := range legacyItems {
		if _, ok := seen[item.ArtifactId]; ok {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func loadResumeDocumentItems(ctx context.Context, db sqlx.SqlConn, userID int64) ([]types.ResumeArtifactItem, error) {
	var rows []resumeArtifactRow
	err := db.QueryRowsCtx(ctx, &rows, `select
d.session_id as artifact_id,
d.title,
d.version,
d.filename,
d.status,
d.chunk_count,
coalesce(s.title, d.title) as bound_session_name,
d.uploaded_at,
d.updated_at
from "public"."resume_documents" d
left join "public"."chat_sessions" s
  on s.session_id = d.session_id and s.user_id = d.user_id
where d.user_id = $1 and d.is_current = true
order by d.updated_at desc, d.id desc`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
			return []types.ResumeArtifactItem{}, nil
		}
		return nil, err
	}

	items := make([]types.ResumeArtifactItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildResumeArtifactItem(row))
	}
	return items, nil
}

func loadLegacyResumeArtifactItems(ctx context.Context, db sqlx.SqlConn, userID int64) ([]types.ResumeArtifactItem, error) {
	var rows []resumeArtifactRow
	err := db.QueryRowsCtx(ctx, &rows, `select
v.chat_id as artifact_id,
coalesce(s.title, v.chat_id) as title,
1::bigint as version,
''::text as filename,
'ready'::text as status,
count(*) as chunk_count,
coalesce(s.title, v.chat_id) as bound_session_name,
max(v.created_at) as uploaded_at,
max(v.created_at) as updated_at
from "public"."vector_store" v
left join "public"."chat_sessions" s
  on s.session_id = v.chat_id and s.user_id = v.user_id
where v.user_id = $1 and v.doc_type = 'resume'
group by v.chat_id, s.title
order by max(v.created_at) desc`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlx.ErrNotFound) {
			return []types.ResumeArtifactItem{}, nil
		}
		return nil, err
	}

	items := make([]types.ResumeArtifactItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildResumeArtifactItem(row))
	}
	return items, nil
}

func loadResumeArtifactItem(ctx context.Context, db sqlx.SqlConn, userID int64, artifactID string) (types.ResumeArtifactItem, error) {
	item, err := loadResumeDocumentItem(ctx, db, userID, artifactID)
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sqlx.ErrNotFound) {
		return types.ResumeArtifactItem{}, err
	}

	return loadLegacyResumeArtifactItem(ctx, db, userID, artifactID)
}

func loadResumeDocumentItem(ctx context.Context, db sqlx.SqlConn, userID int64, artifactID string) (types.ResumeArtifactItem, error) {
	var row resumeArtifactRow
	err := db.QueryRowCtx(ctx, &row, `select
d.session_id as artifact_id,
d.title,
d.version,
d.filename,
d.status,
d.chunk_count,
coalesce(s.title, d.title) as bound_session_name,
d.uploaded_at,
d.updated_at
from "public"."resume_documents" d
left join "public"."chat_sessions" s
  on s.session_id = d.session_id and s.user_id = d.user_id
where d.user_id = $1 and d.session_id = $2 and d.is_current = true
limit 1`, userID, artifactID)
	if err != nil {
		return types.ResumeArtifactItem{}, err
	}
	return buildResumeArtifactItem(row), nil
}

func loadLegacyResumeArtifactItem(ctx context.Context, db sqlx.SqlConn, userID int64, artifactID string) (types.ResumeArtifactItem, error) {
	var row resumeArtifactRow
	err := db.QueryRowCtx(ctx, &row, `select
v.chat_id as artifact_id,
coalesce(s.title, v.chat_id) as title,
1::bigint as version,
''::text as filename,
'ready'::text as status,
count(*) as chunk_count,
coalesce(s.title, v.chat_id) as bound_session_name,
max(v.created_at) as uploaded_at,
max(v.created_at) as updated_at
from "public"."vector_store" v
left join "public"."chat_sessions" s
  on s.session_id = v.chat_id and s.user_id = v.user_id
where v.user_id = $1 and v.chat_id = $2 and v.doc_type = 'resume'
group by v.chat_id, s.title`, userID, artifactID)
	if err != nil {
		return types.ResumeArtifactItem{}, err
	}
	return buildResumeArtifactItem(row), nil
}

func buildResumeArtifactItem(row resumeArtifactRow) types.ResumeArtifactItem {
	return types.ResumeArtifactItem{
		ArtifactId:       row.ArtifactId,
		Title:            row.Title,
		Version:          row.Version,
		Filename:         row.Filename,
		Status:           row.Status,
		ChunkCount:       row.ChunkCount,
		BoundSessionId:   row.ArtifactId,
		BoundSessionName: row.BoundSessionName,
		UploadedAt:       row.UploadedAt.Format(timeLayout),
		UpdatedAt:        row.UpdatedAt.Format(timeLayout),
	}
}
