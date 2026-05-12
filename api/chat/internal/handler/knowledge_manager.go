package handler

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	chatAuth "GoZero-AI/api/chat/internal/auth"
	"GoZero-AI/api/chat/internal/logic"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
)

func KnowledgeDocumentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := optionalKnowledgeViewerContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeDocumentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeDocumentsLogic(ctx, svcCtx)
		resp, err := l.KnowledgeDocuments(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeFoldersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := requireKnowledgeUserContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeFoldersReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeFoldersLogic(ctx, svcCtx)
		resp, err := l.KnowledgeFolders(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeCreateFolderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := requireKnowledgeUserContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeCreateFolderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeCreateFolderLogic(ctx, svcCtx)
		resp, err := l.KnowledgeCreateFolder(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeUpdateFolderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := requireKnowledgeUserContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeUpdateFolderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeUpdateFolderLogic(ctx, svcCtx)
		resp, err := l.KnowledgeUpdateFolder(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeDeleteFolderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := requireKnowledgeUserContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeDeleteFolderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeDeleteFolderLogic(ctx, svcCtx)
		if err := l.KnowledgeDeleteFolder(&req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func KnowledgeMoveDocumentFolderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := requireKnowledgeUserContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeMoveDocumentFolderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeMoveDocumentFolderLogic(ctx, svcCtx)
		resp, err := l.KnowledgeMoveDocumentFolder(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeDocumentChunksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := optionalKnowledgeViewerContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeDocumentChunksReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeDocumentChunksLogic(ctx, svcCtx)
		resp, err := l.KnowledgeDocumentChunks(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func KnowledgeTestQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, ok := optionalKnowledgeViewerContext(w, r, svcCtx)
		if !ok {
			return
		}

		var req types.KnowledgeTestQueryReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		l := logic.NewKnowledgeTestQueryLogic(ctx, svcCtx)
		resp, err := l.KnowledgeTestQuery(&req)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
			return
		}

		httpx.OkJsonCtx(ctx, w, resp)
	}
}

func requireKnowledgeUserContext(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) (context.Context, bool) {
	ctx := r.Context()
	accessToken := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if accessToken == "" {
		httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, map[string]any{
			"message": "请先登录后操作知识库",
		})
		return nil, false
	}

	userID, err := chatAuth.ParseAccessTokenUserID(svcCtx.Config.Auth.AccessSecret, accessToken)
	if err != nil {
		httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, map[string]any{
			"message": "access token 无效或已过期",
		})
		return nil, false
	}

	return chatAuth.WithUserID(ctx, userID), true
}

func optionalKnowledgeViewerContext(w http.ResponseWriter, r *http.Request, svcCtx *svc.ServiceContext) (context.Context, bool) {
	ctx := r.Context()
	accessToken := bearerTokenFromHeader(r.Header.Get("Authorization"))
	if accessToken == "" {
		return ctx, true
	}

	userID, err := chatAuth.ParseAccessTokenUserID(svcCtx.Config.Auth.AccessSecret, accessToken)
	if err != nil {
		httpx.WriteJsonCtx(ctx, w, http.StatusUnauthorized, map[string]any{
			"message": "access token 无效或已过期",
		})
		return nil, false
	}

	return chatAuth.WithUserID(ctx, userID), true
}
