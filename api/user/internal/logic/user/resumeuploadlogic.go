package user

import (
	"context"
	"net/http"
	"strings"

	"GoZero-AI/api/user/internal/svc"
	"GoZero-AI/api/user/internal/types"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/statuserr"

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

	chatID := strings.TrimSpace(req.ChatId)
	if chatID == "" {
		return nil, statuserr.New(http.StatusBadRequest, "chatId 不能为空")
	}

	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return nil, statuserr.New(http.StatusBadRequest, "PDF 未解析出有效文本")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = filename
	}
	modeKey := sessionmode.NormalizeKey(req.Mode)

	chunks := splitText(trimmedContent, l.svcCtx.Config.ResumeChunkSize())
	version, err := l.svcCtx.ResumeStore.SaveResume(l.ctx, userID, chatID, title, filename, modeKey, chunks)
	if err != nil {
		return nil, err
	}

	return &types.ResumeUploadResp{
		Msg:        "私有简历上传成功",
		ChatId:     chatID,
		ArtifactId: chatID,
		Title:      title,
		Filename:   filename,
		Version:    version,
		Chunks:     len(chunks),
	}, nil
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
