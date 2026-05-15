package handler

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"GoZero-AI/api/chat/internal/logic"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
)

func KnowledgeTextUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := requirePublicKnowledgeAdmin(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeTextUploadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewKnowledgeUploadLogic(r.Context(), svcCtx)
		resp, err := l.KnowledgeUpload(&types.KnowledgeUploadInput{
			Title:   strings.TrimSpace(req.Title),
			Content: buildToolKnowledgeContent(req.Source, req.Content),
			Source:  strings.TrimSpace(req.Source),
			UserID:  userID,
		})
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func buildToolKnowledgeContent(source, content string) string {
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		return ""
	}

	trimmedSource := strings.TrimSpace(source)
	if trimmedSource == "" {
		return trimmedContent
	}

	return "资料来源: " + trimmedSource + "\n\n" + trimmedContent
}
