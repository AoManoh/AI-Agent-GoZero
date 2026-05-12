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

type KnowledgeFoldersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeCreateFolderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeUpdateFolderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeDeleteFolderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type KnowledgeMoveDocumentFolderLogic struct {
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

func NewKnowledgeFoldersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeFoldersLogic {
	return &KnowledgeFoldersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeCreateFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeCreateFolderLogic {
	return &KnowledgeCreateFolderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeUpdateFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeUpdateFolderLogic {
	return &KnowledgeUpdateFolderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeDeleteFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeDeleteFolderLogic {
	return &KnowledgeDeleteFolderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func NewKnowledgeMoveDocumentFolderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeMoveDocumentFolderLogic {
	return &KnowledgeMoveDocumentFolderLogic{
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
	filter, err := l.buildKnowledgeScopeFilter(userID, req.Visibility, req.FolderScoped, req.FolderId)
	if err != nil {
		return nil, err
	}
	filter.Limit = limit

	documents, err := l.svcCtx.VectorStore.ListKnowledgeDocumentsFiltered(l.ctx, filter)
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

func (l *KnowledgeFoldersLogic) KnowledgeFolders(*types.KnowledgeFoldersReq) (*types.KnowledgeFoldersResp, error) {
	userID, err := requiredKnowledgeUserID(l.ctx)
	if err != nil {
		return nil, err
	}

	folders, err := l.svcCtx.VectorStore.ListKnowledgeFolders(l.ctx, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("知识库目录暂不可用，请稍后重试")
	}
	unfiledCount, err := l.svcCtx.VectorStore.CountUnfiledKnowledgeDocuments(l.ctx, userID)
	if err != nil {
		return nil, statuserr.ServiceUnavailable("知识库目录暂不可用，请稍后重试")
	}

	items := buildKnowledgeFolderTree(folders)

	return &types.KnowledgeFoldersResp{
		Folders:      items,
		UnfiledCount: unfiledCount,
		Total:        int64(len(folders)),
		TotalCount:   int64(len(folders)),
		Initialized:  false,
		Meta:         buildKnowledgeManagerMeta(&userID),
	}, nil
}

func (l *KnowledgeCreateFolderLogic) KnowledgeCreateFolder(req *types.KnowledgeCreateFolderReq) (*types.KnowledgeFolderMutationResp, error) {
	userID, err := requiredKnowledgeUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if err := validateKnowledgeFolderName(req.Name); err != nil {
		return nil, err
	}
	if req.ParentId < 0 {
		return nil, statuserr.New(http.StatusBadRequest, "父目录 ID 无效")
	}

	folder, err := l.svcCtx.VectorStore.CreateKnowledgeFolder(l.ctx, svc.KnowledgeCreateFolderInput{
		UserID:    userID,
		Name:      req.Name,
		ParentID:  req.ParentId,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return nil, mapKnowledgeMutationError(err, "父目录不存在或无权访问")
	}

	return &types.KnowledgeFolderMutationResp{
		Folder: buildKnowledgeFolderItem(folder),
		Meta:   buildKnowledgeManagerMeta(&userID),
	}, nil
}

func (l *KnowledgeUpdateFolderLogic) KnowledgeUpdateFolder(req *types.KnowledgeUpdateFolderReq) (*types.KnowledgeFolderMutationResp, error) {
	userID, err := requiredKnowledgeUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识目录 ID 无效")
	}
	if strings.TrimSpace(req.Name) != "" {
		if err := validateKnowledgeFolderName(req.Name); err != nil {
			return nil, err
		}
	}
	if req.ParentId < 0 {
		return nil, statuserr.New(http.StatusBadRequest, "父目录 ID 无效")
	}

	folder, err := l.svcCtx.VectorStore.UpdateKnowledgeFolder(l.ctx, svc.KnowledgeUpdateFolderInput{
		UserID:       userID,
		ID:           req.Id,
		Name:         req.Name,
		ParentID:     req.ParentId,
		SortOrder:    req.SortOrder,
		SetParent:    req.SetParent,
		SetSortOrder: req.SetSortOrder,
	})
	if err != nil {
		return nil, mapKnowledgeMutationError(err, "知识目录不存在或无权访问")
	}

	return &types.KnowledgeFolderMutationResp{
		Folder: buildKnowledgeFolderItem(folder),
		Meta:   buildKnowledgeManagerMeta(&userID),
	}, nil
}

func (l *KnowledgeDeleteFolderLogic) KnowledgeDeleteFolder(req *types.KnowledgeDeleteFolderReq) (*types.KnowledgeFolderDeleteResp, error) {
	userID, err := requiredKnowledgeUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识目录 ID 无效")
	}

	if err := l.svcCtx.VectorStore.DeleteKnowledgeFolder(l.ctx, req.Id, userID); err != nil {
		return nil, mapKnowledgeMutationError(err, "知识目录不存在或无权访问")
	}

	return &types.KnowledgeFolderDeleteResp{
		Deleted: true,
		Meta:    buildKnowledgeManagerMeta(&userID),
	}, nil
}

func (l *KnowledgeMoveDocumentFolderLogic) KnowledgeMoveDocumentFolder(req *types.KnowledgeMoveDocumentFolderReq) (*types.KnowledgeDocumentMutationResp, error) {
	userID, err := requiredKnowledgeUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if req.Id <= 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识库文档 ID 无效")
	}
	if req.FolderId < 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识目录 ID 无效")
	}

	document, err := l.svcCtx.VectorStore.MoveKnowledgeDocumentFolder(l.ctx, req.Id, userID, req.FolderId)
	if err != nil {
		return nil, mapKnowledgeMutationError(err, "知识库文档或目录不存在，或无权访问")
	}

	return &types.KnowledgeDocumentMutationResp{
		Document: buildKnowledgeDocumentItem(document, l.svcCtx.Config.EmbeddingModel()),
		Meta:     buildKnowledgeManagerMeta(&userID),
	}, nil
}

func (l *KnowledgeDocumentChunksLogic) KnowledgeDocumentChunks(req *types.KnowledgeDocumentChunksReq) (*types.KnowledgeDocumentChunksResp, error) {
	if req.Id <= 0 {
		return nil, statuserr.New(http.StatusBadRequest, "知识库文档 ID 无效")
	}

	userID := optionalKnowledgeUserID(l.ctx)
	limit := boundedKnowledgeLimit(req.Limit, 50, 500)
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
	filter, err := l.buildKnowledgeScopeFilter(userID, req.Visibility, req.FolderScoped, req.FolderId)
	if err != nil {
		return nil, err
	}
	results, err := l.svcCtx.VectorStore.SearchKnowledgeFiltered(l.ctx, query, topK, filter)
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

func requiredKnowledgeUserID(ctx context.Context) (int64, error) {
	if userID, ok := chatAuth.UserIDFromContext(ctx); ok && userID > 0 {
		return userID, nil
	}

	return 0, statuserr.Unauthorized("请先登录后操作知识库")
}

func buildKnowledgeManagerMeta(userID *int64) types.KnowledgeManagerMeta {
	scope := "public"
	if userID != nil {
		scope = "public+private"
	}

	return types.KnowledgeManagerMeta{
		SchemaVersion: "knowledge-manager-v2",
		Available:     true,
		Scope:         scope,
		GeneratedAt:   time.Now().Format(chatTimeLayout),
	}
}

func (l *KnowledgeDocumentsLogic) buildKnowledgeScopeFilter(userID *int64, visibility string, folderScoped bool, folderID int64) (svc.KnowledgeScopeFilter, error) {
	return buildKnowledgeScopeFilter(l.ctx, l.svcCtx, userID, visibility, folderScoped, folderID)
}

func (l *KnowledgeTestQueryLogic) buildKnowledgeScopeFilter(userID *int64, visibility string, folderScoped bool, folderID int64) (svc.KnowledgeScopeFilter, error) {
	return buildKnowledgeScopeFilter(l.ctx, l.svcCtx, userID, visibility, folderScoped, folderID)
}

func buildKnowledgeScopeFilter(ctx context.Context, svcCtx *svc.ServiceContext, userID *int64, visibility string, folderScoped bool, folderID int64) (svc.KnowledgeScopeFilter, error) {
	normalizedVisibility := strings.ToLower(strings.TrimSpace(visibility))
	if normalizedVisibility != "" && normalizedVisibility != "public" && normalizedVisibility != "private" {
		return svc.KnowledgeScopeFilter{}, statuserr.New(http.StatusBadRequest, "知识可见性无效")
	}
	if folderID < 0 {
		return svc.KnowledgeScopeFilter{}, statuserr.New(http.StatusBadRequest, "知识目录 ID 无效")
	}
	if folderScoped {
		if userID == nil {
			return svc.KnowledgeScopeFilter{}, statuserr.Unauthorized("请先登录后访问目录知识")
		}
		if normalizedVisibility != "" && normalizedVisibility != "private" {
			return svc.KnowledgeScopeFilter{}, statuserr.New(http.StatusBadRequest, "目录范围仅支持私人知识")
		}
		if folderID > 0 {
			if _, err := svcCtx.VectorStore.LoadKnowledgeFolder(ctx, folderID, *userID); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return svc.KnowledgeScopeFilter{}, statuserr.NotFound("知识目录不存在或无权访问")
				}
				return svc.KnowledgeScopeFilter{}, statuserr.ServiceUnavailable("知识目录暂不可用，请稍后重试")
			}
		}
		normalizedVisibility = "private"
	}
	if normalizedVisibility == "private" && userID == nil {
		return svc.KnowledgeScopeFilter{}, statuserr.Unauthorized("请先登录后访问私人知识")
	}

	return svc.KnowledgeScopeFilter{
		UserID:       userID,
		Visibility:   normalizedVisibility,
		FolderScoped: folderScoped,
		FolderID:     folderID,
	}, nil
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

	item := types.KnowledgeDocumentItem{
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
	if document.FolderID.Valid {
		item.FolderId = document.FolderID.Int64
	}

	return item
}

func buildKnowledgeFolderItem(folder svc.KnowledgeFolder) types.KnowledgeFolderItem {
	item := types.KnowledgeFolderItem{
		Id:            folder.ID,
		Name:          folder.Name,
		Path:          folder.Path,
		Depth:         folder.Depth,
		SortOrder:     folder.SortOrder,
		DocumentCount: folder.DocumentCount,
		ChunkCount:    folder.ChunkCount,
		CreatedAt:     folder.CreatedAt.Format(chatTimeLayout),
		UpdatedAt:     folder.UpdatedAt.Format(chatTimeLayout),
	}
	if folder.ParentID.Valid {
		item.ParentId = folder.ParentID.Int64
	}

	return item
}

func buildKnowledgeFolderTree(folders []svc.KnowledgeFolder) []types.KnowledgeFolderItem {
	childrenByParent := make(map[int64][]svc.KnowledgeFolder, len(folders))
	roots := make([]svc.KnowledgeFolder, 0)
	for _, folder := range folders {
		if folder.ParentID.Valid {
			childrenByParent[folder.ParentID.Int64] = append(childrenByParent[folder.ParentID.Int64], folder)
			continue
		}
		roots = append(roots, folder)
	}

	var buildNode func(folder svc.KnowledgeFolder) types.KnowledgeFolderItem
	buildNode = func(folder svc.KnowledgeFolder) types.KnowledgeFolderItem {
		item := buildKnowledgeFolderItem(folder)
		for _, child := range childrenByParent[folder.ID] {
			item.Children = append(item.Children, buildNode(child))
		}
		return item
	}

	items := make([]types.KnowledgeFolderItem, 0, len(roots))
	for _, root := range roots {
		items = append(items, buildNode(root))
	}
	return items
}

func validateKnowledgeFolderName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return statuserr.New(http.StatusBadRequest, "目录名称不能为空")
	}
	if len([]rune(trimmed)) > 80 {
		return statuserr.New(http.StatusBadRequest, "目录名称不能超过 80 个字符")
	}

	return nil
}

func mapKnowledgeMutationError(err error, notFoundMessage string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return statuserr.NotFound(notFoundMessage)
	}
	if _, ok := statuserr.StatusCode(err); ok {
		return err
	}

	return statuserr.ServiceUnavailable("知识库操作暂不可用，请稍后重试")
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
