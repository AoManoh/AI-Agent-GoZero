package logic

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/zeromicro/go-zero/core/logx"

	chatAuth "GoZero-AI/api/chat/internal/auth"
	"GoZero-AI/api/chat/internal/svc"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/internal/statuserr"
)

// 知识库相关常量
const chatTimeLayout = "2006-01-02T15:04:05Z07:00"

// knowledgeEmbeddingDimension 是知识库向量维度的快照常量。
//
// 与 db/user.sql 中 knowledge_base.embedding 列的 vector(1536) 类型严格对齐；
// 切到不同维度的 embedding 模型时需要同步迁移 DDL + reindex 全部历史数据，
// 不在本次范围（v0.3+ 可观察性升级或独立的 RAG 模型升级 task）。
// 2026-05-12 Q8=B 决策：暴露给前端卡片元信息展示，拒绝产品图的 "1024d" 设计稿示意。
const knowledgeEmbeddingDimension = 1536

type KnowledgeDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeDocumentChunksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeTestQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeDocumentsLogic {
	return &KnowledgeDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeDocumentChunksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeDocumentChunksLogic {
	return &KnowledgeDocumentChunksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeTestQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeTestQueryLogic {
	return &KnowledgeTestQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowledgeDocumentsLogic) KnowledgeDocuments(req *types.KnowledgeDocumentsReq) (*types.KnowledgeDocumentsResp, error) {
	userID := optionalKnowledgeUserID(l.ctx)
	limit := boundedKnowledgeLimit(req.Limit, 20, 100)

	documents, err := l.svcCtx.VectorStore.ListKnowledgeDocuments(l.ctx, userID, limit)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("知识库文档暂不可用，请稍后重试")
	}

	embeddingModel := l.svcCtx.Config.EmbeddingModel()
	items := make([]types.KnowledgeDocumentItem, 0, len(documents))
	for _, doc := range documents {
		items = append(items, buildKnowledgeDocumentItem(doc, embeddingModel))
	}

	return &types.KnowledgeDocumentsResp{
		Documents: items,
		Total:     int64(len(items)),
		Meta:      buildKnowledgeManagerMeta(userID),
	}, nil
}

func (l *KnowledgeDocumentChunksLogic) KnowledgeDocumentChunks(req *types.KnowledgeDocumentChunksReq) (*types.KnowledgeDocumentChunksResp, error) {
	if req.Id <= 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识库文档 ID 无效")
	}

	userID := optionalKnowledgeUserID(l.ctx)
	limit := boundedKnowledgeLimit(req.Limit, 50, 200)
	document, chunks, err := l.svcCtx.VectorStore.LoadKnowledgeDocumentChunks(l.ctx, req.Id, userID, limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, statuserr.NotFound("知识库文档不存在或无权访问")
		}
		return nil, statuserr.ServiceUnavailable("知识库文档分块暂不可用，请稍后重试")
	}

	items := make([]types.KnowledgeChunkItem, 0, len(chunks))
	for _, chunk := range chunks {
		items = append(items, types.KnowledgeChunkItem{
			ChunkId:   chunk.ID,
			Title:     chunk.Title,
			Content:   truncateKnowledgeContent(chunk.Content, 900),
			CreatedAt: chunk.CreatedAt.Format(chatTimeLayout),
		})
	}

	return &types.KnowledgeDocumentChunksResp{
		Document: buildKnowledgeDocumentItem(document, l.svcCtx.Config.EmbeddingModel()),
		Chunks:   items,
		Meta:     buildKnowledgeManagerMeta(userID),
	}, nil
}

func (l *KnowledgeTestQueryLogic) KnowledgeTestQuery(req *types.KnowledgeTestQueryReq) (*types.KnowledgeTestQueryResp, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, statuserr.New(http.StatusBadRequest, "测试查询内容不能为空")
	}

	userID := optionalKnowledgeUserID(l.ctx)
	topK := boundedKnowledgeLimit(req.TopK, 5, 10)
	results, err := l.svcCtx.VectorStore.SearchKnowledge(l.ctx, query, topK, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("知识库测试召回暂不可用，请稍后重试")
	}

	items := make([]types.KnowledgeChunkItem, 0, len(results))
	for _, result := range results {
		items = append(items, types.KnowledgeChunkItem{
			ChunkId: result.ID,
			Title:   result.Title,
			Content: truncateKnowledgeContent(result.Content, 900),
			Score:   result.Score,
		})
	}

	return &types.KnowledgeTestQueryResp{
		Results: items,
		Total:   int64(len(items)),
		Meta:    buildKnowledgeManagerMeta(userID),
	}, nil
}

func optionalKnowledgeUserID(ctx context.Context) *int64 {
	if userID, ok := chatAuth.UserIDFromContext(ctx); ok {
		return &userID
	}

	return nil
}

func buildKnowledgeManagerMeta(userID *int64) types.KnowledgeManagerMeta {
	scope := "public"
	if userID != nil {
		scope = "public+private"
	}

	return types.KnowledgeManagerMeta{
		SchemaVersion: "knowledge-manager-v1",
		Available:     true,
		Scope:         scope,
		GeneratedAt:   time.Now().Format(chatTimeLayout),
	}
}

// buildKnowledgeDocumentItem 把后端 svc 层的 KnowledgeDocument 聚合数据转换为 types 层 API 响应项。
//
// 2026-05-12 Q8=B 派生：
//   - SizeBytes 来自 svc 层 SUM(LENGTH(content)) 聚合
//   - EmbeddingDimension 使用 knowledgeEmbeddingDimension 常量（与 DDL vector(1536) 对齐）
//   - EmbeddingModel 由调用方从 svcCtx.Config.EmbeddingModel() 取出后传入，避免本函数对 svcCtx 形成耦合
func buildKnowledgeDocumentItem(document svc.KnowledgeDocument, embeddingModel string) types.KnowledgeDocumentItem {
	scope := document.Visibility
	if scope == "" && document.OwnerID == 1 {
		scope = "public"
	}
	if scope == "" {
		scope = "private"
	}

	return types.KnowledgeDocumentItem{
		DocumentId:         document.DocumentID,
		Title:              document.Title,
		Scope:              scope,
		Source:             document.Source,
		Visibility:         document.Visibility,
		Status:             document.Status,
		Version:            document.Version,
		OwnerId:            document.OwnerID,
		ChunkCount:         document.ChunkCount,
		SizeBytes:          document.SizeBytes,
		EmbeddingDimension: knowledgeEmbeddingDimension,
		EmbeddingModel:     embeddingModel,
		Preview:            truncateKnowledgeContent(document.Preview, 180),
		CreatedAt:          document.CreatedAt.Format(chatTimeLayout),
		UpdatedAt:          document.UpdatedAt.Format(chatTimeLayout),
	}
}

func boundedKnowledgeLimit(value, fallback, max int64) int {
	if fallback <= 0 {
		fallback = 20
	}
	if max < fallback {
		max = fallback
	}
	if value <= 0 {
		return int(fallback)
	}
	if value > max {
		return int(max)
	}

	return int(value)
}

func truncateKnowledgeContent(content string, maxRunes int) string {
	trimmed := strings.TrimSpace(content)
	if maxRunes <= 0 {
		return trimmed
	}

	runes := []rune(trimmed)
	if len(runes) <= maxRunes {
		return trimmed
	}

	return string(runes[:maxRunes]) + "..."
}
