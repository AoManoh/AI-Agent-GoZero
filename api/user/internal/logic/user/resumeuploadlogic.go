package user

import (
	"context"
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/statuserr"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type ResumeUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResumeUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeUploadLogic {
	return &ResumeUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResumeUploadLogic) ResumeUpload(req *types.ResumeUploadReq, filename, content string) (*types.ResumeUploadResp, error) {
	userID, err := currentUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	legacySessionID := strings.TrimSpace(req.ChatId)
	artifactID := legacySessionID
	if artifactID == "" {
		artifactID = newResumeArtifactID()
	}

	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, statuserr.Coded(http.StatusBadRequest, "empty_text", "PDF 未解析出有效文本")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = filename
	}
	modeKey := sessionmode.NormalizeKey(req.Mode)

	chunks := splitText(trimmedContent, l.svcCtx.Config.ResumeChunkSize())
	version, err := l.svcCtx.ResumeStore.SaveResume(l.ctx, userID, artifactID, legacySessionID, title, filename, modeKey, chunks)
	if err != nil {
		return nil, err
	}

	return &types.ResumeUploadResp{
		Msg:             "私有简历上传成功",
		ChatId:          legacySessionID,
		LegacySessionId: legacySessionID,
		ArtifactId:      artifactID,
		Title:           title,
		Filename:        filename,
		Version:         version,
		Status:          "ready",
		Chunks:          len(chunks),
		ParseStatus:     buildReadyResumeParseStatus(int64(len(chunks))),
	}, nil
}

func newResumeArtifactID() string {
	return "resume_" + uuid.NewString()
}

func buildReadyResumeParseStatus(chunks int64) types.ResumeParseStatus {
	return types.ResumeParseStatus{
		Stage:           "ready",
		Progress:        100,
		TotalChunks:     chunks,
		ProcessedChunks: chunks,
		FailedChunks:    0,
		Retryable:       false,
	}
}

func splitText(text string, maxChunkSize int) []string {
	var chunks []string
	if text == "" {
		return chunks
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
