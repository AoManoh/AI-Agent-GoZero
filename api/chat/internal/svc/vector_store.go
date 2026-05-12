// Package svc 提供服务上下文和依赖管理功能
// vector_store.go 文件实现了基于 PostgreSQL + pgvector 扩展的向量数据库存储层
//
// 主要功能:
//  1. 向量化消息存储 - 将用户对话和 AI 响应转换为向量并持久化存储
//  2. 语义相似度搜索 - 基于向量相似度检索历史对话和相关知识
//  3. RAG 知识库管理 - 支持文档分块、向量化和检索功能
//  4. 会话持久化 - 替代内存存储，支持分布式部署和故障恢复
//
// 核心组件:
//   - VectorStore: 向量存储核心结构，实现 SessionStore 接口
//   - PostgreSQL + pgvector: 提供高性能向量存储和检索能力
//   - OpenAI Embedding API: 将文本转换为高维向量表示
//
// 技术特性:
//   - 支持多字节字符(中文、emoji等)的完整处理
//   - 基于连接池的高并发数据库访问
//   - 向量相似度计算和排序
//   - 会话状态管理和历史对话检索
//
// 应用场景:
//   - AI 对话系统的上下文记忆
//   - 企业知识库的语义搜索
//   - 智能客服的历史对话分析
//   - RAG(检索增强生成)系统的知识检索
package svc

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/sessionruntime"
	"GoZero-AI/internal/statuserr"

	// pgvector-go 是一个 Go 语言库，用于将 Go 语言中的数据类型转换为 PostgreSQL 的 pgvector 扩展所需的格式
	"github.com/pgvector/pgvector-go"
)

// VectorStore 向量存储结构
type VectorStore struct {
	Pool           postgresPool   // 数据库连接池
	OpenAIClient   *openai.Client // OpenAI客户端
	EmbeddingModel string         // 向量模型名称
}

type postgresPool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Close()
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type SessionInterviewConfig struct {
	DirectionKey          string
	DirectionLabel        string
	DifficultyLevel       int64
	DifficultyLabel       string
	InterviewerStyle      string
	InterviewerStyleLabel string
	FocusAreas            []byte
	FollowUpDepth         string
	EstimatedMinutes      int64
	ProgressPercent       int64
	ResumeArtifactID      string
}

type SessionPracticeContext struct {
	QuestionKey      string
	Source           string
	QuestionSnapshot string
}

type SessionRuntimeContext struct {
	ScenarioType       string
	StarterSource      string
	StarterQuestionKey string
	QuestionKey        string
	Source             string
	QuestionSnapshot   string
}

type KnowledgeDocument struct {
	DocumentID int64
	OwnerID    int64
	FolderID   sql.NullInt64
	Title      string
	Source     string
	Visibility string
	Status     string
	Version    int64
	Preview    string
	ChunkCount int64
	SizeBytes  int64 // 2026-05-12 Q8=B：所有 chunk content 字节数求和（SUM(LENGTH(content))），用于前端卡片元信息展示
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type KnowledgeDocumentChunk struct {
	ID        int64
	Title     string
	Content   string
	CreatedAt time.Time
}

type KnowledgeSearchResult struct {
	ID      int64
	Title   string
	Content string
	Score   float64
}

type KnowledgeFolder struct {
	ID            int64
	UserID        int64
	ParentID      sql.NullInt64
	Name          string
	Path          string
	Depth         int64
	SortOrder     int64
	DocumentCount int64
	ChunkCount    int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type KnowledgeDeleteFolderResult struct {
	MovedDocCount       int64
	PromotedFolderCount int64
	ParentName          string
}

type KnowledgeScopeFilter struct {
	UserID       *int64
	Limit        int
	Visibility   string
	FolderScoped bool
	FolderID     int64
}

type KnowledgeCreateFolderInput struct {
	UserID    int64
	Name      string
	ParentID  int64
	SortOrder int64
}

type KnowledgeUpdateFolderInput struct {
	UserID       int64
	ID           int64
	Name         string
	ParentID     int64
	SortOrder    int64
	SetParent    bool
	SetSortOrder bool
}

type execQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const vectorStoreEmbeddingDimensions = 1536

var (
	ErrSessionDeleted      = statuserr.Conflict("会话已删除，请创建新会话")
	ErrSessionCompleted    = statuserr.Conflict("面试已结束，请创建新会话")
	ErrSessionAccessDenied = statuserr.Forbidden("无权访问该会话")
)

// NewVectorStore 初始化向量存储
func NewVectorStore(cfg config.VectorDBConfig, openAIClient *openai.Client) (*VectorStore, error) {
	// 1. 构建连接字符串, 使用 pgxpool 连接数据库
	// 这里会通过读取配置文件中（api\internal\config\config.go）的信息
	// 来构建连接字符串，连接到数据库
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// 2. 解析配置
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(cfg.MaxConn) // 设置最大连接数

	// 配置亚洲时区
	if cfg.TimeZone != "" {
		if poolConfig.ConnConfig.RuntimeParams == nil {
			poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
		}
		poolConfig.ConnConfig.RuntimeParams["TimeZone"] = cfg.TimeZone
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			if _, err := conn.Exec(ctx, fmt.Sprintf("SET TIME ZONE '%s'", cfg.TimeZone)); err != nil {
				return fmt.Errorf("设置数据库时区失败: %w", err)
			}
			return nil
		}
	}

	// 3. 创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	// 4. 填充 VectorStore 结构体并返回
	return &VectorStore{
		Pool:           pool,
		OpenAIClient:   openAIClient,
		EmbeddingModel: cfg.EmbeddingModel,
	}, nil
}

// TestConnection 实现数据库连接测试
func (vs *VectorStore) TestConnection() error {
	// 创建一个上下文对象，设置一个超时时间，如果超时则返回错误
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 使用 Ping 方法测试连接
	return vs.Pool.Ping(ctx)
}

// generateEmbedding 清洗 content 文本，并将其转换为 pgvector.Vector 类型的向量
func (vs *VectorStore) generateEmbedding(ctx context.Context, content string) (pgvector.Vector, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if content == "" {
		return pgvector.NewVector(nil), errors.New("内容不能为空")
	}

	// 使用 OpenAI 客户端生成向量, 返回一个 EmbeddingResponse 结构体
	// 这个结构体包含了一个 Data 字段，其中 Data 字段是一个 Embedding 结构体的切片
	// 每个 Embedding 结构体包含一个 Embedding 字段，这个字段是一个 []float32 类型的向量
	embedding, err := vs.OpenAIClient.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(vs.EmbeddingModel), // 使用配置文件中的模型名称
			Input: []string{content},                        // 用户对话的上下文内容
		},
	)
	if err != nil {
		return pgvector.NewVector(nil), fmt.Errorf("OpenAI API 错误，生成向量失败: %w", err)
	}

	if len(embedding.Data) == 0 {
		return pgvector.NewVector(nil), errors.New("OpenAI API 错误，生成向量失败: 没有返回数据")
	}

	// 将向量转换为 pgvector.vector 格式
	// 返回向量数据
	return pgvector.NewVector(embedding.Data[0].Embedding), nil
}

// SaveMessage 保存消息到向量数据库
// 新增 RAG 本地知识库后，进行了升级
func (vs *VectorStore) SaveMessage(chatId, role, content string) error {
	return vs.SaveMessageWithUser(context.Background(), chatId, role, content, nil, "")
}

func (vs *VectorStore) SaveMessageWithUser(ctx context.Context, chatId, role, content string, userID *int64, mode string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := vs.ValidateSessionWrite(ctx, chatId, userID); err != nil {
		return err
	}

	// 步骤1：生成  pgvector.vector  文本向量
	embedding, err := vs.generateEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %w", err)
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if userID != nil {
		if err := ensureSessionWritable(ctx, tx, chatId, *userID); err != nil {
			return err
		}
	}

	// 步骤2：数据库存储
	sql := `INSERT INTO vector_store (chat_id, user_id, role, content, embedding, doc_type) 
            VALUES ($1, $2, $3, $4, $5, 'message')`

	_, err = tx.Exec(ctx, sql, chatId, nullableUserID(userID), role, content, embedding)
	if err != nil {
		return err
	}

	if err := ensureChatSession(ctx, tx, chatId, userID, role, content, mode); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (vs *VectorStore) SaveMessageBodyWithUser(ctx context.Context, chatId, role, content string, userID *int64, mode string) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := vs.ValidateSessionWrite(ctx, chatId, userID); err != nil {
		return 0, err
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer rollbackTx(tx)

	if userID != nil {
		if err := ensureSessionWritable(ctx, tx, chatId, *userID); err != nil {
			return 0, err
		}
	}

	var messageID int64
	zeroEmbedding := zeroMessageEmbedding()
	insertSQL := `INSERT INTO vector_store (chat_id, user_id, role, content, embedding, doc_type)
VALUES ($1, $2, $3, $4, $5, 'message')
RETURNING id`
	if err := tx.QueryRow(ctx, insertSQL, chatId, nullableUserID(userID), role, content, zeroEmbedding).Scan(&messageID); err != nil {
		return 0, err
	}

	if err := ensureChatSession(ctx, tx, chatId, userID, role, content, mode); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return messageID, nil
}

func (vs *VectorStore) UpdateMessageEmbedding(ctx context.Context, messageID int64, content string) error {
	if messageID <= 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	embedding, err := vs.generateEmbedding(ctx, content)
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %w", err)
	}

	_, err = vs.Pool.Exec(ctx, `UPDATE vector_store
SET embedding = $1
WHERE id = $2 AND doc_type = 'message'`, embedding, messageID)
	return err
}

func zeroMessageEmbedding() pgvector.Vector {
	return pgvector.NewVector(make([]float32, vectorStoreEmbeddingDimensions))
}

func (vs *VectorStore) ValidateSessionWrite(ctx context.Context, chatId string, userID *int64) error {
	if userID == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return ensureSessionWritable(ctx, vs.Pool, chatId, *userID)
}

// GetMessage 获取消息从向量数据库
func (vs *VectorStore) GetMessage(chatId string, limit int) ([]types.VectorMessage, error) {
	return vs.GetMessageWithUser(chatId, nil, limit)
}

func (vs *VectorStore) GetMessageWithUser(chatId string, userID *int64, limit int) ([]types.VectorMessage, error) {
	return vs.GetMessageWithUserContext(context.Background(), chatId, userID, limit)
}

func (vs *VectorStore) GetMessageWithUserContext(ctx context.Context, chatId string, userID *int64, limit int) ([]types.VectorMessage, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. 构建一条 SQL 语句，从表中获取消息
	var (
		sql  string
		args []any
	)
	if userID == nil {
		sql = `SELECT role, content FROM vector_store WHERE chat_id = $1 AND user_id IS NULL AND doc_type = 'message' ORDER BY created_at DESC LIMIT $2`
		args = []any{chatId, limit}
	} else {
		sql = `SELECT role, content FROM vector_store WHERE chat_id = $1 AND user_id = $2 AND doc_type = 'message' ORDER BY created_at DESC LIMIT $3`
		args = []any{chatId, *userID, limit}
	}

	// 2. 传入指定用户 ID，获取他的历史对话数据，并做出限制，我们肯定不希望数据太多导致准确性降低
	rows, err := vs.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("从数据库中获取当前用户消息失败: %w", err)
	}
	defer rows.Close()

	// 3. 遍历结果集，将每一行数据转换为 VectorMessage 结构体
	var messages []types.VectorMessage
	for rows.Next() {
		// 3.1 定义两个变量，用于存储每一行数据中的 role 和 content 字段
		var role, content string
		// 3.2 使用 rows.Scan 方法将每一行数据中的 role 和 content 字段赋值给变量
		if err := rows.Scan(&role, &content); err != nil {
			return nil, fmt.Errorf("行扫描数据失败: %w", err)
		}
		// 3.3 将每一行数据转换为 VectorMessage 结构体，并添加到 messages 切片中
		messages = append(messages, types.VectorMessage{
			Role:    role,
			Content: content,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取当前用户消息失败: %w", err)
	}

	// 4. 反转 messages 切片，因为数据库中的数据是按创建时间降序排列的
	// 用户肯定希望看到最新的消息，所以需要反转切片
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// SaveKnowledge 保存知识块到知识库
// 用于构建RAG系统的知识底座，支持文档分块存储和语义检索
//
// **后续**调用模式:
//   - KnowledgeUploadLogic: 循环调用保存PDF分块
//   - 调用顺序: PDF上传 -> 文本提取 -> Logic层分块 -> 逐块调用SaveKnowledge
//   - 与RetrieveKnowledge()配合实现完整的"存储-检索"循环
//
// 参数说明:
//
//	title: 知识文档的标题(如PDF文件名)，用于结果展示和溯源
//	chunk: 文档的一个分块片段，长度应控制在合理范围内
func (vs *VectorStore) SaveKnowledge(title, chunk string) error {
	return vs.SaveKnowledgeForUser(context.Background(), title, chunk, 1)
}

func (vs *VectorStore) SaveKnowledgeForUser(ctx context.Context, title, chunk string, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	return vs.SaveKnowledgeBatchForUserContextWithMeta(ctx, title, []string{chunk}, userID, "")
}

// RetrieveKnowledge 基于语义相似度检索知识库
// RAG系统的核心检索方法，为用户问题找到最相关的知识内容
//
// 技术原理:
//  1. 查询向量化 - 将用户问题转换为向量表示
//  2. 向量相似度计算 - 使用PostgreSQL的向量操作符<->计算余弦距离
//  3. TopK排序 - 按相似度排序并返回最相关的K个结果
//  4. 结果封装 - 返回结构化的知识块信息，不包含向量数据
//
// **后续**使用场景:
//   - ChatLogic.Chat(): 生成AI回复前检索相关知识
//   - 调用时机: 用户提问 -> 检索知识 -> 注入AI上下文 -> 生成回答
//   - 与SaveKnowledge()共同构建完整RAG闭环
//
// 参数说明:
//
//	query: 用户的问题或查询语句，将被向量化用于相似度匹配
//	topK: 返回的最大结果数量，控制检索精度和性能平衡
//
// 返回值:
//
//	[]types.KnowledgeChunk: 按相似度排序的知识块列表
//	error: 向量生成或数据库检索错误
func (vs *VectorStore) RetrieveKnowledge(query string, topK int) ([]types.KnowledgeChunk, error) {
	return vs.RetrieveKnowledgeScoped(query, topK, nil, "")
}

func (vs *VectorStore) RetrieveKnowledgeScoped(query string, topK int, userID *int64, chatID string) ([]types.KnowledgeChunk, error) {
	return vs.RetrieveKnowledgeScopedContext(context.Background(), query, topK, userID, chatID)
}

func (vs *VectorStore) RetrieveKnowledgeScopedContext(ctx context.Context, query string, topK int, userID *int64, chatID string) ([]types.KnowledgeChunk, error) {
	return vs.retrieveKnowledgeScopedContext(ctx, query, topK, userID, chatID, true)
}

func (vs *VectorStore) RetrievePrivateSessionKnowledgeScopedContext(ctx context.Context, query string, topK int, userID *int64, chatID string) ([]types.KnowledgeChunk, error) {
	resumeScopeID := strings.TrimSpace(chatID)
	if userID != nil && resumeScopeID != "" {
		if artifactID, ok := vs.loadSessionResumeArtifactID(ctx, *userID, resumeScopeID); ok {
			resumeScopeID = artifactID
		}
	}
	return vs.retrieveKnowledgeScopedContext(ctx, query, topK, userID, resumeScopeID, false)
}

func (vs *VectorStore) retrieveKnowledgeScopedContext(ctx context.Context, query string, topK int, userID *int64, chatID string, includePublic bool) ([]types.KnowledgeChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 1. 将用户查询转换为向量表示
	// 使用相同的embedding模型确保向量空间一致性
	queryEmbeddingVector, err := vs.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("用户查询向量化失败: %w", err)
	}

	var results []scoredKnowledgeChunk

	if includePublic {
		publicResults, err := vs.fetchPublicKnowledge(ctx, queryEmbeddingVector, topK, userID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			// 当前环境可能仍保留 jsonb/旧维度知识库；这里降级到私有简历与消息链路，避免整条对话失败。
			fmt.Printf("skip public knowledge retrieval due to schema/runtime mismatch: %v\n", err)
		} else {
			results = append(results, publicResults...)
		}
	}

	if userID != nil && chatID != "" {
		resumeResults, err := vs.fetchResumeKnowledge(ctx, queryEmbeddingVector, topK, *userID, chatID)
		if err != nil {
			return nil, fmt.Errorf("私有简历检索失败: %w", err)
		}
		results = append(results, resumeResults...)
	}

	sortScoredKnowledge(results)

	if len(results) > topK {
		results = results[:topK]
	}

	knowledge := make([]types.KnowledgeChunk, 0, len(results))
	for _, result := range results {
		knowledge = append(knowledge, result.KnowledgeChunk)
	}

	return knowledge, nil
}

func (vs *VectorStore) ListKnowledgeFolders(ctx context.Context, userID int64) ([]KnowledgeFolder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if userID <= 0 {
		return nil, fmt.Errorf("目录所属用户不能为空")
	}

	rows, err := vs.Pool.Query(ctx, `SELECT
f.id,
f.user_id,
f.parent_id,
f.name,
f.path,
f.depth,
f.sort_order,
count(DISTINCT (kb.title, kb.source, kb.version)) AS document_count,
count(kb.id) AS chunk_count,
f.created_at,
f.updated_at
FROM knowledge_folders f
LEFT JOIN knowledge_base kb
  ON kb.folder_id = f.id
 AND kb.user_id = f.user_id
 AND kb.visibility = 'private'
 AND kb.status = 'ready'
WHERE f.user_id = $1
GROUP BY f.id, f.user_id, f.parent_id, f.name, f.path, f.depth, f.sort_order, f.created_at, f.updated_at
ORDER BY f.parent_id NULLS FIRST, f.sort_order ASC, f.created_at ASC, f.id ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := make([]KnowledgeFolder, 0)
	for rows.Next() {
		var folder KnowledgeFolder
		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.ParentID,
			&folder.Name,
			&folder.Path,
			&folder.Depth,
			&folder.SortOrder,
			&folder.DocumentCount,
			&folder.ChunkCount,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

func (vs *VectorStore) CreateKnowledgeFolder(ctx context.Context, input KnowledgeCreateFolderInput) (KnowledgeFolder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	name := strings.TrimSpace(input.Name)
	if err := validateKnowledgeFolderMutation(input.UserID, 0, name); err != nil {
		return KnowledgeFolder{}, err
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return KnowledgeFolder{}, err
	}
	defer rollbackTx(tx)

	if input.ParentID > 0 {
		if err := ensureKnowledgeFolderOwned(ctx, tx, input.ParentID, input.UserID); err != nil {
			return KnowledgeFolder{}, err
		}
	}

	parentPath, parentDepth, err := loadKnowledgeFolderParentPath(ctx, tx, input.ParentID, input.UserID)
	if err != nil {
		return KnowledgeFolder{}, err
	}
	depth := parentDepth
	if input.ParentID > 0 {
		depth = parentDepth + 1
	}
	if depth > 2 {
		return KnowledgeFolder{}, statuserr.New(http.StatusUnprocessableEntity, "目录最多支持 3 层")
	}
	path := buildKnowledgeFolderPath(parentPath, name)

	row := tx.QueryRow(ctx, `INSERT INTO knowledge_folders (user_id, parent_id, name, path, depth, sort_order)
VALUES ($1, NULLIF($2::bigint, 0), $3, $4, $5, $6)
RETURNING id, user_id, parent_id, name, path, depth, sort_order, 0::bigint, 0::bigint, created_at, updated_at`,
		input.UserID,
		input.ParentID,
		name,
		path,
		depth,
		input.SortOrder,
	)

	folder, err := scanKnowledgeFolder(row)
	if err != nil {
		return KnowledgeFolder{}, translateKnowledgeFolderWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return KnowledgeFolder{}, err
	}

	return folder, nil
}

func (vs *VectorStore) UpdateKnowledgeFolder(ctx context.Context, input KnowledgeUpdateFolderInput) (KnowledgeFolder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	name := strings.TrimSpace(input.Name)
	if err := validateKnowledgeFolderMutation(input.UserID, input.ID, name); err != nil {
		return KnowledgeFolder{}, err
	}
	if name == "" && !input.SetParent && !input.SetSortOrder {
		return vs.LoadKnowledgeFolder(ctx, input.ID, input.UserID)
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return KnowledgeFolder{}, err
	}
	defer rollbackTx(tx)

	if err := ensureKnowledgeFolderOwned(ctx, tx, input.ID, input.UserID); err != nil {
		return KnowledgeFolder{}, err
	}
	if input.SetParent {
		if input.ParentID == input.ID {
			return KnowledgeFolder{}, fmt.Errorf("目录不能移动到自身下")
		}
		if input.ParentID > 0 {
			if err := ensureKnowledgeFolderOwned(ctx, tx, input.ParentID, input.UserID); err != nil {
				return KnowledgeFolder{}, err
			}
			if err := ensureKnowledgeFolderNotDescendant(ctx, tx, input.ID, input.UserID, input.ParentID); err != nil {
				return KnowledgeFolder{}, err
			}
		}
	}

	row := tx.QueryRow(ctx, `UPDATE knowledge_folders
SET name = CASE WHEN $3 THEN $4 ELSE name END,
    parent_id = CASE WHEN $5 THEN NULLIF($6::bigint, 0) ELSE parent_id END,
    sort_order = CASE WHEN $7 THEN $8 ELSE sort_order END,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, parent_id, name, sort_order, 0::bigint, 0::bigint, created_at, updated_at`,
		input.ID,
		input.UserID,
		name != "",
		name,
		input.SetParent,
		input.ParentID,
		input.SetSortOrder,
		input.SortOrder,
	)

	folder, err := scanKnowledgeFolder(row)
	if err != nil {
		return KnowledgeFolder{}, translateKnowledgeFolderWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return KnowledgeFolder{}, err
	}

	return folder, nil
}

func (vs *VectorStore) LoadKnowledgeFolder(ctx context.Context, folderID, userID int64) (KnowledgeFolder, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if folderID <= 0 || userID <= 0 {
		return KnowledgeFolder{}, pgx.ErrNoRows
	}

	return scanKnowledgeFolder(vs.Pool.QueryRow(ctx, `SELECT
f.id,
f.user_id,
f.parent_id,
f.name,
f.sort_order,
count(DISTINCT (kb.title, kb.source, kb.version)) AS document_count,
count(kb.id) AS chunk_count,
f.created_at,
f.updated_at
FROM knowledge_folders f
LEFT JOIN knowledge_base kb
  ON kb.folder_id = f.id
 AND kb.user_id = f.user_id
 AND kb.visibility = 'private'
 AND kb.status = 'ready'
WHERE f.id = $1 AND f.user_id = $2
GROUP BY f.id, f.user_id, f.parent_id, f.name, f.sort_order, f.created_at, f.updated_at`,
		folderID,
		userID,
	))
}

func (vs *VectorStore) DeleteKnowledgeFolder(ctx context.Context, folderID, userID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if folderID <= 0 || userID <= 0 {
		return pgx.ErrNoRows
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackTx(tx)

	if err := ensureKnowledgeFolderOwned(ctx, tx, folderID, userID); err != nil {
		return err
	}

	var childCount int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM knowledge_folders WHERE user_id = $1 AND parent_id = $2`, userID, folderID).Scan(&childCount); err != nil {
		return err
	}
	if childCount > 0 {
		return statuserr.Conflict("目录下仍有子目录，不能删除")
	}

	var chunkCount int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM knowledge_base WHERE user_id = $1 AND folder_id = $2`, userID, folderID).Scan(&chunkCount); err != nil {
		return err
	}
	if chunkCount > 0 {
		return statuserr.Conflict("目录下仍有文档，不能删除")
	}

	tag, err := tx.Exec(ctx, `DELETE FROM knowledge_folders WHERE id = $1 AND user_id = $2`, folderID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return tx.Commit(ctx)
}

func (vs *VectorStore) MoveKnowledgeDocumentFolder(ctx context.Context, documentID, userID, folderID int64) (KnowledgeDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if documentID <= 0 || userID <= 0 {
		return KnowledgeDocument{}, pgx.ErrNoRows
	}

	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return KnowledgeDocument{}, err
	}
	defer rollbackTx(tx)

	if folderID > 0 {
		if err := ensureKnowledgeFolderOwned(ctx, tx, folderID, userID); err != nil {
			return KnowledgeDocument{}, err
		}
	}

	var title, source string
	var version int64
	if err := tx.QueryRow(ctx, `SELECT title, coalesce(source, ''), coalesce(version, 1)
FROM knowledge_base
WHERE id = $1 AND user_id = $2 AND visibility = 'private'
LIMIT 1`, documentID, userID).Scan(&title, &source, &version); err != nil {
		return KnowledgeDocument{}, err
	}

	tag, err := tx.Exec(ctx, `UPDATE knowledge_base
SET folder_id = NULLIF($5::bigint, 0),
    updated_at = now()
WHERE user_id = $1 AND title = $2 AND source = $3 AND version = $4 AND visibility = 'private'`,
		userID,
		title,
		source,
		version,
		folderID,
	)
	if err != nil {
		return KnowledgeDocument{}, err
	}
	if tag.RowsAffected() == 0 {
		return KnowledgeDocument{}, pgx.ErrNoRows
	}

	if err := tx.Commit(ctx); err != nil {
		return KnowledgeDocument{}, err
	}

	return vs.LoadKnowledgeDocument(ctx, documentID, &userID)
}

func (vs *VectorStore) ListKnowledgeDocuments(ctx context.Context, userID *int64, limit int) ([]KnowledgeDocument, error) {
	return vs.ListKnowledgeDocumentsFiltered(ctx, KnowledgeScopeFilter{
		UserID: userID,
		Limit:  limit,
	})
}

func (vs *VectorStore) ListKnowledgeDocumentsFiltered(ctx context.Context, filter KnowledgeScopeFilter) ([]KnowledgeDocument, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	whereSQL, args, err := knowledgeDocumentFilterWhere(filter, 1, false)
	if err != nil {
		return nil, err
	}
	// 2026-05-12 Q8=B：聚合每个文档的总字节数（sum(length(content))），暴露给前端卡片元信息
	query := fmt.Sprintf(`SELECT
min(id) AS document_id,
user_id,
coalesce(max(folder_id), 0) AS folder_id,
title,
coalesce(source, '') AS source,
coalesce(max(visibility), case when user_id = %d then 'public' else 'private' end) AS visibility,
coalesce(max(status), 'ready') AS status,
coalesce(max(version), 1) AS version,
count(*) AS chunk_count,
coalesce(sum(length(content)), 0) AS size_bytes,
min(created_at) AS created_at,
coalesce(max(updated_at), max(created_at)) AS updated_at,
left(coalesce((array_agg(content ORDER BY created_at ASC, id ASC))[1], ''), 240) AS preview
FROM knowledge_base
WHERE %s
GROUP BY user_id, title, source, version
ORDER BY coalesce(max(updated_at), max(created_at)) DESC, min(id) DESC
LIMIT $%d`, publicKnowledgeOwnerID, whereSQL, len(args)+1)
	args = append(args, filter.Limit)

	rows, err := vs.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make([]KnowledgeDocument, 0)
	for rows.Next() {
		var doc KnowledgeDocument
		var folderID int64
		if err := rows.Scan(&doc.DocumentID, &doc.OwnerID, &folderID, &doc.Title, &doc.Source, &doc.Visibility, &doc.Status, &doc.Version, &doc.ChunkCount, &doc.SizeBytes, &doc.CreatedAt, &doc.UpdatedAt, &doc.Preview); err != nil {
			return nil, err
		}
		if folderID > 0 {
			doc.FolderID = sql.NullInt64{Int64: folderID, Valid: true}
		}
		documents = append(documents, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

func (vs *VectorStore) LoadKnowledgeDocument(ctx context.Context, documentID int64, userID *int64) (KnowledgeDocument, error) {
	document, _, err := vs.LoadKnowledgeDocumentChunks(ctx, documentID, userID, 1)
	return document, err
}

func (vs *VectorStore) LoadKnowledgeDocumentChunks(ctx context.Context, documentID int64, userID *int64, limit int) (KnowledgeDocument, []KnowledgeDocumentChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		limit = 50
	}

	whereSQL, args := knowledgeDocumentAccessWhere(userID, 2)
	args = append([]any{documentID}, args...)
	row := vs.Pool.QueryRow(ctx, fmt.Sprintf(`SELECT user_id, title, coalesce(source, ''), coalesce(version, 1)
FROM knowledge_base
WHERE id = $1 AND %s
LIMIT 1`, whereSQL), args...)

	var (
		ownerID int64
		title   string
		source  string
		version int64
	)
	if err := row.Scan(&ownerID, &title, &source, &version); err != nil {
		return KnowledgeDocument{}, nil, err
	}

	var (
		document         KnowledgeDocument
		documentFolderID int64
	)
	// 2026-05-12 Q8=B：聚合每个文档的总字节数（sum(length(content))），与 ListKnowledgeDocuments 保持一致
	if err := vs.Pool.QueryRow(ctx, `SELECT
min(id) AS document_id,
user_id,
coalesce(max(folder_id), 0) AS folder_id,
title,
coalesce(source, '') AS source,
coalesce(max(visibility), case when user_id = $1 then 'public' else 'private' end) AS visibility,
coalesce(max(status), 'ready') AS status,
coalesce(max(version), 1) AS version,
count(*) AS chunk_count,
coalesce(sum(length(content)), 0) AS size_bytes,
min(created_at) AS created_at,
coalesce(max(updated_at), max(created_at)) AS updated_at,
left(coalesce((array_agg(content ORDER BY created_at ASC, id ASC))[1], ''), 240) AS preview
FROM knowledge_base
WHERE user_id = $1 AND title = $2 AND source = $3 AND version = $4
GROUP BY user_id, title, source, version`, ownerID, title, source, version).Scan(
		&document.DocumentID,
		&document.OwnerID,
		&documentFolderID,
		&document.Title,
		&document.Source,
		&document.Visibility,
		&document.Status,
		&document.Version,
		&document.ChunkCount,
		&document.SizeBytes,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.Preview,
	); err != nil {
		return KnowledgeDocument{}, nil, err
	}
	if documentFolderID > 0 {
		document.FolderID = sql.NullInt64{Int64: documentFolderID, Valid: true}
	}

	rows, err := vs.Pool.Query(ctx, `SELECT id, title, content, created_at
FROM knowledge_base
WHERE user_id = $1 AND title = $2 AND source = $3 AND version = $4
ORDER BY created_at ASC, id ASC
LIMIT $5`, ownerID, title, source, version, limit)
	if err != nil {
		return KnowledgeDocument{}, nil, err
	}
	defer rows.Close()

	chunks := make([]KnowledgeDocumentChunk, 0)
	for rows.Next() {
		var chunk KnowledgeDocumentChunk
		if err := rows.Scan(&chunk.ID, &chunk.Title, &chunk.Content, &chunk.CreatedAt); err != nil {
			return KnowledgeDocument{}, nil, err
		}
		chunks = append(chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return KnowledgeDocument{}, nil, err
	}

	return document, chunks, nil
}

func (vs *VectorStore) SearchKnowledge(ctx context.Context, query string, topK int, userID *int64) ([]KnowledgeSearchResult, error) {
	return vs.SearchKnowledgeFiltered(ctx, query, topK, KnowledgeScopeFilter{
		UserID: userID,
	})
}

func (vs *VectorStore) SearchKnowledgeFiltered(ctx context.Context, query string, topK int, filter KnowledgeScopeFilter) ([]KnowledgeSearchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if topK <= 0 {
		topK = 5
	}

	queryEmbeddingVector, err := vs.generateEmbedding(ctx, strings.TrimSpace(query))
	if err != nil {
		return nil, fmt.Errorf("知识库测试查询向量化失败: %w", err)
	}

	scoredResults, err := vs.fetchKnowledgeForManager(ctx, queryEmbeddingVector, topK, filter)
	if err != nil {
		return nil, err
	}
	sortScoredKnowledge(scoredResults)
	if len(scoredResults) > topK {
		scoredResults = scoredResults[:topK]
	}

	results := make([]KnowledgeSearchResult, 0, len(scoredResults))
	for _, item := range scoredResults {
		results = append(results, KnowledgeSearchResult{
			ID:      item.ID,
			Title:   item.Title,
			Content: item.Content,
			Score:   item.Score,
		})
	}

	return results, nil
}

func ensureChatSession(ctx context.Context, db execQuerier, chatId string, userID *int64, role, content, mode string) error {
	if userID == nil {
		return nil
	}

	title := defaultSessionTitle
	if role == openai.ChatMessageRoleUser {
		title = deriveSessionTitle(content)
	}
	modeKey := sessionmode.NormalizeKey(mode)

	query := `INSERT INTO chat_sessions (session_id, user_id, title, mode, last_message_at, message_count, is_active)
VALUES ($1, $2, $3, $4, now(), 1, true)
ON CONFLICT (session_id) DO UPDATE
SET user_id = COALESCE(chat_sessions.user_id, EXCLUDED.user_id),
    title = CASE
        WHEN chat_sessions.title = $5 AND EXCLUDED.title <> '' THEN EXCLUDED.title
        ELSE chat_sessions.title
    END,
    mode = CASE
        WHEN chat_sessions.mode IS NULL OR btrim(chat_sessions.mode) = '' THEN EXCLUDED.mode
        WHEN chat_sessions.mode = $6 AND EXCLUDED.mode <> $6 THEN EXCLUDED.mode
        ELSE chat_sessions.mode
    END,
    updated_at = now(),
    last_message_at = now(),
    message_count = chat_sessions.message_count + 1`

	_, err := db.Exec(ctx, query, chatId, *userID, title, modeKey, defaultSessionTitle, defaultSessionMode)
	return err
}

func (vs *VectorStore) ResolveSessionMode(ctx context.Context, chatId string, userID *int64, requestedMode string) (string, error) {
	if userID == nil || strings.TrimSpace(chatId) == "" {
		return effectiveSessionMode("", requestedMode), nil
	}

	var storedMode string
	row := vs.Pool.QueryRow(ctx, `SELECT mode FROM chat_sessions WHERE session_id = $1 AND user_id = $2 LIMIT 1`, chatId, *userID)
	if err := row.Scan(&storedMode); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return effectiveSessionMode("", requestedMode), nil
		}
		return "", err
	}

	return effectiveSessionMode(storedMode, requestedMode), nil
}

func (vs *VectorStore) LoadSessionInterviewConfig(ctx context.Context, chatId string, userID *int64) (SessionInterviewConfig, bool, error) {
	if userID == nil || strings.TrimSpace(chatId) == "" {
		return SessionInterviewConfig{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var config SessionInterviewConfig
	row := vs.Pool.QueryRow(ctx, `SELECT
direction_key,
direction_label,
difficulty_level,
difficulty_label,
interviewer_style,
interviewer_style_label,
focus_areas,
follow_up_depth,
estimated_minutes,
progress_percent,
resume_artifact_id
FROM chat_sessions
WHERE session_id = $1 AND user_id = $2 AND is_active = true
LIMIT 1`, chatId, *userID)
	if err := row.Scan(
		&config.DirectionKey,
		&config.DirectionLabel,
		&config.DifficultyLevel,
		&config.DifficultyLabel,
		&config.InterviewerStyle,
		&config.InterviewerStyleLabel,
		&config.FocusAreas,
		&config.FollowUpDepth,
		&config.EstimatedMinutes,
		&config.ProgressPercent,
		&config.ResumeArtifactID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionInterviewConfig{}, false, nil
		}
		return SessionInterviewConfig{}, false, err
	}
	return config, true, nil
}

func (vs *VectorStore) loadSessionResumeArtifactID(ctx context.Context, userID int64, chatID string) (string, bool) {
	if ctx == nil {
		ctx = context.Background()
	}
	var artifactID string
	row := vs.Pool.QueryRow(ctx, `SELECT resume_artifact_id
FROM chat_sessions
WHERE session_id = $1 AND user_id = $2 AND is_active = true
LIMIT 1`, chatID, userID)
	if err := row.Scan(&artifactID); err != nil {
		return "", false
	}
	artifactID = strings.TrimSpace(artifactID)
	if artifactID == "" {
		return "", false
	}
	return artifactID, true
}

func (vs *VectorStore) LoadSessionRuntimeContext(ctx context.Context, chatId string, userID *int64) (SessionRuntimeContext, bool, error) {
	if userID == nil || strings.TrimSpace(chatId) == "" {
		return SessionRuntimeContext{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var runtime SessionRuntimeContext
	row := vs.Pool.QueryRow(ctx, `SELECT scenario_type, starter_source, starter_question_key
FROM chat_sessions
WHERE session_id = $1 AND user_id = $2 AND is_active = true
LIMIT 1`, chatId, *userID)
	if err := row.Scan(&runtime.ScenarioType, &runtime.StarterSource, &runtime.StarterQuestionKey); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionRuntimeContext{}, false, nil
		}
		return SessionRuntimeContext{}, false, err
	}
	runtime.ScenarioType = sessionruntime.NormalizeScenario(runtime.ScenarioType)
	runtime.StarterSource = sessionruntime.NormalizeStarterSource(runtime.StarterSource)
	runtime.StarterQuestionKey = strings.TrimSpace(runtime.StarterQuestionKey)

	if runtime.ScenarioType == sessionruntime.ScenarioQuestionPractice {
		practice, found, err := vs.LoadSessionPracticeContext(ctx, chatId, userID)
		if err != nil {
			return SessionRuntimeContext{}, false, err
		}
		if found {
			runtime.QuestionKey = practice.QuestionKey
			runtime.Source = practice.Source
			runtime.QuestionSnapshot = practice.QuestionSnapshot
		}
		if strings.TrimSpace(runtime.QuestionKey) == "" {
			runtime.QuestionKey = runtime.StarterQuestionKey
		}
	}

	return runtime, true, nil
}

func (vs *VectorStore) LoadSessionPracticeContext(ctx context.Context, chatId string, userID *int64) (SessionPracticeContext, bool, error) {
	if userID == nil || strings.TrimSpace(chatId) == "" {
		return SessionPracticeContext{}, false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var practice SessionPracticeContext
	row := vs.Pool.QueryRow(ctx, `SELECT question_key, source, question_snapshot
FROM session_question_events
WHERE session_id = $1 AND user_id = $2 AND source = 'bank'
ORDER BY turn_index ASC
LIMIT 1`, chatId, *userID)
	if err := row.Scan(&practice.QuestionKey, &practice.Source, &practice.QuestionSnapshot); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionPracticeContext{}, false, nil
		}
		return SessionPracticeContext{}, false, err
	}
	return practice, true, nil
}

func ensureSessionWritable(ctx context.Context, db execQuerier, chatId string, userID int64) error {
	row := db.QueryRow(ctx, `SELECT user_id, is_active, completed_at IS NOT NULL FROM chat_sessions WHERE session_id = $1 LIMIT 1`, chatId)

	var ownerID int64
	var isActive bool
	var isCompleted bool
	if err := row.Scan(&ownerID, &isActive, &isCompleted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if ownerID != userID {
		return ErrSessionAccessDenied
	}
	if isCompleted {
		return ErrSessionCompleted
	}
	if !isActive {
		return ErrSessionDeleted
	}

	return nil
}

func deriveSessionTitle(content string) string {
	const maxLen = 48

	normalized := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(content, "\n", " "), "\r", " "))
	if normalized == "" {
		return defaultSessionTitle
	}
	runes := []rune(normalized)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}

	return normalized
}

func nullableUserID(userID *int64) any {
	if userID == nil {
		return nil
	}

	return *userID
}

func effectiveSessionMode(storedMode, requestedMode string) string {
	if strings.TrimSpace(storedMode) != "" {
		return sessionmode.NormalizeKey(storedMode)
	}

	return sessionmode.NormalizeKey(requestedMode)
}

func knowledgeAccessWhere(userID *int64, startIndex int) (string, []any) {
	return knowledgeAccessWhereWithStatus(userID, startIndex, true)
}

func knowledgeDocumentAccessWhere(userID *int64, startIndex int) (string, []any) {
	return knowledgeAccessWhereWithStatus(userID, startIndex, false)
}

func knowledgeAccessWhereWithStatus(userID *int64, startIndex int, requireReady bool) (string, []any) {
	statusSQL := ""
	if requireReady {
		statusSQL = " AND status = 'ready'"
	}

	if userID == nil || *userID == publicKnowledgeOwnerID {
		return fmt.Sprintf("(visibility = 'public' OR user_id = %d)%s", publicKnowledgeOwnerID, statusSQL), nil
	}

	return fmt.Sprintf("((visibility = 'public' OR user_id = %d) OR user_id = $%d)%s", publicKnowledgeOwnerID, startIndex, statusSQL), []any{*userID}
}

func knowledgeDocumentFilterWhere(filter KnowledgeScopeFilter, startIndex int, requireReady bool) (string, []any, error) {
	if startIndex <= 0 {
		startIndex = 1
	}

	visibility := strings.ToLower(strings.TrimSpace(filter.Visibility))
	if visibility != "" && visibility != "public" && visibility != "private" {
		return "", nil, fmt.Errorf("知识可见性无效")
	}
	if filter.FolderScoped {
		if filter.UserID == nil || *filter.UserID <= 0 {
			return "", nil, fmt.Errorf("目录范围需要登录用户")
		}
		if filter.FolderID < 0 {
			return "", nil, fmt.Errorf("目录 ID 无效")
		}
		if visibility != "" && visibility != "private" {
			return "", nil, fmt.Errorf("目录范围仅支持私人知识")
		}
		visibility = "private"
	}

	args := make([]any, 0, 3)
	nextParam := startIndex
	addArg := func(value any) string {
		args = append(args, value)
		placeholder := fmt.Sprintf("$%d", nextParam)
		nextParam++
		return placeholder
	}

	conditions := make([]string, 0, 5)
	switch visibility {
	case "public":
		conditions = append(conditions, fmt.Sprintf("(visibility = 'public' OR user_id = %d)", publicKnowledgeOwnerID))
	case "private":
		if filter.UserID == nil || *filter.UserID <= 0 {
			return "", nil, fmt.Errorf("私人知识范围需要登录用户")
		}
		conditions = append(conditions, fmt.Sprintf("user_id = %s", addArg(*filter.UserID)))
		conditions = append(conditions, "visibility = 'private'")
	default:
		if filter.UserID == nil || *filter.UserID == publicKnowledgeOwnerID {
			conditions = append(conditions, fmt.Sprintf("(visibility = 'public' OR user_id = %d)", publicKnowledgeOwnerID))
		} else {
			conditions = append(conditions, fmt.Sprintf("((visibility = 'public' OR user_id = %d) OR user_id = %s)", publicKnowledgeOwnerID, addArg(*filter.UserID)))
		}
	}

	if filter.FolderScoped {
		if filter.FolderID == 0 {
			conditions = append(conditions, "folder_id IS NULL")
		} else {
			conditions = append(conditions, fmt.Sprintf("folder_id = %s", addArg(filter.FolderID)))
		}
	}
	if requireReady {
		conditions = append(conditions, "status = 'ready'")
	}

	return strings.Join(conditions, " AND "), args, nil
}

const (
	defaultSessionTitle    = "新对话"
	defaultSessionMode     = sessionmode.DefaultKey
	publicKnowledgeOwnerID = 1
)

type scoredKnowledgeChunk struct {
	types.KnowledgeChunk
	Score float64
}

func (vs *VectorStore) fetchKnowledgeForManager(ctx context.Context, queryEmbeddingVector pgvector.Vector, topK int, filter KnowledgeScopeFilter) ([]scoredKnowledgeChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	whereSQL, args, err := knowledgeDocumentFilterWhere(filter, 2, true)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT id, title, content, embedding <-> $1 AS score
FROM knowledge_base
WHERE %s AND embedding IS NOT NULL
ORDER BY score
LIMIT $%d`, whereSQL, len(args)+2)
	args = append([]any{queryEmbeddingVector}, args...)
	args = append(args, topK)

	rows, err := vs.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scoredKnowledgeChunk
	for rows.Next() {
		var (
			id      int64
			title   string
			content string
			score   float64
		)
		if err := rows.Scan(&id, &title, &content, &score); err != nil {
			return nil, err
		}
		results = append(results, scoredKnowledgeChunk{
			KnowledgeChunk: types.KnowledgeChunk{
				ID:      id,
				Title:   title,
				Content: content,
			},
			Score: score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (vs *VectorStore) fetchPublicKnowledge(ctx context.Context, queryEmbeddingVector pgvector.Vector, topK int, userID *int64) ([]scoredKnowledgeChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sql := `SELECT id, title, content, embedding <-> $1 AS score
FROM knowledge_base
WHERE ((visibility = 'public' OR user_id = 1)`
	args := []any{queryEmbeddingVector}
	if userID != nil {
		sql += ` OR user_id = $2`
		args = append(args, *userID)
	}
	sql += `) AND status = 'ready' AND embedding IS NOT NULL`
	sql += `
ORDER BY score
LIMIT $`
	sql += fmt.Sprintf("%d", len(args)+1)
	args = append(args, topK)

	rows, err := vs.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scoredKnowledgeChunk
	for rows.Next() {
		var (
			id      int64
			title   string
			content string
			score   float64
		)
		if err := rows.Scan(&id, &title, &content, &score); err != nil {
			return nil, err
		}
		results = append(results, scoredKnowledgeChunk{
			KnowledgeChunk: types.KnowledgeChunk{
				ID:      id,
				Title:   title,
				Content: content,
			},
			Score: score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (vs *VectorStore) fetchResumeKnowledge(ctx context.Context, queryEmbeddingVector pgvector.Vector, topK int, userID int64, chatID string) ([]scoredKnowledgeChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	rows, err := vs.Pool.Query(ctx, `SELECT id, '[resume]'::text AS title, content, embedding <-> $1 AS score
FROM vector_store
WHERE user_id = $2 AND chat_id = $3 AND doc_type = 'resume'
ORDER BY score
LIMIT $4`, queryEmbeddingVector, userID, chatID, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scoredKnowledgeChunk
	for rows.Next() {
		var (
			id      int64
			title   string
			content string
			score   float64
		)
		if err := rows.Scan(&id, &title, &content, &score); err != nil {
			return nil, err
		}
		results = append(results, scoredKnowledgeChunk{
			KnowledgeChunk: types.KnowledgeChunk{
				ID:      id,
				Title:   title,
				Content: content,
			},
			Score: score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func sortScoredKnowledge(results []scoredKnowledgeChunk) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score < results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func scanKnowledgeFolder(row pgx.Row) (KnowledgeFolder, error) {
	var folder KnowledgeFolder
	if err := row.Scan(
		&folder.ID,
		&folder.UserID,
		&folder.ParentID,
		&folder.Name,
		&folder.SortOrder,
		&folder.DocumentCount,
		&folder.ChunkCount,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	); err != nil {
		return KnowledgeFolder{}, err
	}

	return folder, nil
}

func validateKnowledgeFolderMutation(userID, folderID int64, name string) error {
	if userID <= 0 {
		return fmt.Errorf("目录所属用户不能为空")
	}
	if folderID < 0 {
		return fmt.Errorf("目录 ID 无效")
	}
	if folderID == 0 && name == "" {
		return fmt.Errorf("目录名称不能为空")
	}
	if name != "" && len([]rune(name)) > 120 {
		return fmt.Errorf("目录名称不能超过 120 个字符")
	}

	return nil
}

func ensureKnowledgeFolderOwned(ctx context.Context, db execQuerier, folderID, userID int64) error {
	if folderID <= 0 || userID <= 0 {
		return pgx.ErrNoRows
	}

	var id int64
	if err := db.QueryRow(ctx, `SELECT id FROM knowledge_folders WHERE id = $1 AND user_id = $2 LIMIT 1`, folderID, userID).Scan(&id); err != nil {
		return err
	}

	return nil
}

func ensureKnowledgeFolderNotDescendant(ctx context.Context, db execQuerier, folderID, userID, parentID int64) error {
	var isDescendant bool
	if err := db.QueryRow(ctx, `WITH RECURSIVE descendants AS (
    SELECT id FROM knowledge_folders WHERE parent_id = $1 AND user_id = $2
    UNION ALL
    SELECT f.id
    FROM knowledge_folders f
    INNER JOIN descendants d ON f.parent_id = d.id
    WHERE f.user_id = $2
)
SELECT EXISTS(SELECT 1 FROM descendants WHERE id = $3)`, folderID, userID, parentID).Scan(&isDescendant); err != nil {
		return err
	}
	if isDescendant {
		return statuserr.Conflict("目录不能移动到自己的子目录下")
	}

	return nil
}

func translateKnowledgeFolderWriteError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return statuserr.Conflict("同级目录下已存在同名目录")
		case "23514":
			return statuserr.New(http.StatusBadRequest, "目录名称不能为空")
		}
	}

	return err
}

// SaveKnowledgeBatch 以事务方式批量写入知识块，确保全部成功或全部失败
// SaveKnowledgeBatch 批量保存知识块到向量数据库
// 参数:
//   - title: 知识块的标题
//   - chunks: 需要保存的知识块内容切片
//
// 返回值:
//   - error: 操作过程中遇到的错误，如果成功则为nil
func (vs *VectorStore) SaveKnowledgeBatch(title string, chunks []string) error {
	return vs.SaveKnowledgeBatchForUser(title, chunks, 1)
}

func (vs *VectorStore) SaveKnowledgeBatchForUser(title string, chunks []string, userID int64) error {
	return vs.SaveKnowledgeBatchForUserContext(context.Background(), title, chunks, userID)
}

func (vs *VectorStore) SaveKnowledgeBatchForUserContext(ctx context.Context, title string, chunks []string, userID int64) error {
	return vs.SaveKnowledgeBatchForUserContextWithMeta(ctx, title, chunks, userID, "")
}

func (vs *VectorStore) SaveKnowledgeBatchForUserContextWithMeta(ctx context.Context, title string, chunks []string, userID int64, source string) error {
	return vs.SaveKnowledgeBatchForUserContextWithMetaInFolder(ctx, title, chunks, userID, source, nil)
}

func (vs *VectorStore) SaveKnowledgeBatchForUserContextWithMetaInFolder(ctx context.Context, title string, chunks []string, userID int64, source string, folderID *int64) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// 如果知识块切片为空，直接返回nil
	if len(chunks) == 0 {
		return nil
	}
	if userID <= 0 {
		return fmt.Errorf("知识所有者不能为空")
	}

	type knowledgeEmbedding struct {
		chunk     string
		embedding pgvector.Vector
	}

	embeddings := make([]knowledgeEmbedding, 0, len(chunks))
	for _, chunk := range chunks {
		embedding, embedErr := vs.generateEmbedding(ctx, chunk)
		if embedErr != nil {
			return fmt.Errorf("知识块向量化失败: %w", embedErr)
		}
		embeddings = append(embeddings, knowledgeEmbedding{
			chunk:     chunk,
			embedding: embedding,
		})
	}

	// 开启数据库事务
	tx, err := vs.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("创建事务失败: %w", err)
	}

	// 使用独立短超时上下文回滚，避免请求取消后事务清理被同一个 ctx 跳过。
	defer rollbackTx(tx)

	knowledgeSource := defaultKnowledgeSource(source, title)
	visibility := knowledgeVisibilityForUser(userID)
	var knowledgeFolderID any
	if folderID != nil {
		if *folderID <= 0 {
			return fmt.Errorf("知识目录 ID 无效")
		}
		if visibility != "private" {
			return statuserr.New(http.StatusBadRequest, "公共知识暂不支持目录归档")
		}
		if err := ensureKnowledgeFolderOwned(ctx, tx, *folderID, userID); err != nil {
			return err
		}
		knowledgeFolderID = *folderID
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, knowledgeDocumentLockKey(userID, title, knowledgeSource)); err != nil {
		return fmt.Errorf("锁定知识文档身份失败: %w", err)
	}

	var version int64
	if err := tx.QueryRow(ctx, `SELECT coalesce(max(version), 0) + 1
FROM knowledge_base
WHERE user_id = $1 AND title = $2 AND source = $3`, userID, title, knowledgeSource).Scan(&version); err != nil {
		return fmt.Errorf("计算知识文档版本失败: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE knowledge_base
SET status = 'archived', updated_at = now()
WHERE user_id = $1 AND title = $2 AND source = $3 AND status = 'ready'`, userID, title, knowledgeSource); err != nil {
		return fmt.Errorf("归档旧知识文档版本失败: %w", err)
	}

	// 定义插入 SQL 语句
	insertSQL := `INSERT INTO knowledge_base (title, content, embedding, user_id, folder_id, source, visibility, status, version, content_hash, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'ready', $8, $9, now())`
	// 遍历所有知识块
	for _, item := range embeddings {
		// 执行 SQL 插入语句
		if _, err := tx.Exec(ctx, insertSQL, title, item.chunk, item.embedding, userID, knowledgeFolderID, knowledgeSource, visibility, version, knowledgeContentHash(item.chunk)); err != nil {
			return fmt.Errorf("保存知识块到数据库失败: %w", err)
		}
	}

	// 提交事务
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("提交事务失败: %w", commitErr)
	}

	return nil
}

func defaultKnowledgeSource(source, title string) string {
	trimmedSource := strings.TrimSpace(source)
	if trimmedSource != "" {
		return trimmedSource
	}
	return strings.TrimSpace(title)
}

func knowledgeVisibilityForUser(userID int64) string {
	if userID == publicKnowledgeOwnerID {
		return "public"
	}
	return "private"
}

func knowledgeContentHash(content string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(content)))
	return hex.EncodeToString(sum[:])
}

func knowledgeDocumentLockKey(userID int64, title, source string) string {
	return fmt.Sprintf("%d:%s:%s", userID, strings.TrimSpace(title), strings.TrimSpace(source))
}

func rollbackTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
