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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/types"
	"GoZero-AI/internal/sessionmode"
	"GoZero-AI/internal/statuserr"
	// pgvector-go 是一个 Go 语言库，用于将 Go 语言中的数据类型转换为 PostgreSQL 的 pgvector 扩展所需的格式
	"github.com/pgvector/pgvector-go"
)

// VectorStore 向量存储结构
type VectorStore struct {
	Pool           *pgxpool.Pool  // 数据库连接池
	OpenAIClient   *openai.Client // OpenAI客户端
	EmbeddingModel string         // 向量模型名称
}

type execQuerier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	ErrSessionDeleted      = statuserr.Conflict("会话已删除，请创建新会话")
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
	if userID <= 0 {
		return fmt.Errorf("知识所有者不能为空")
	}

	// 1. 将知识块内容转换为向量表示
	// generateEmbedding()返回已序列化的向量数据
	embeddingVector, err := vs.generateEmbedding(ctx, chunk)
	if err != nil {
		return fmt.Errorf("知识块向量化失败: %w", err)
	}

	// 2. 将知识块和向量数据存储到knowledge_base表
	// 与vector_store表分离，专门存储知识库数据
	sql := `INSERT INTO knowledge_base (title, content, embedding, user_id) VALUES ($1, $2, $3, $4)`
	_, err = vs.Pool.Exec(ctx, sql, title, chunk, embeddingVector, userID)
	if err != nil {
		return fmt.Errorf("保存知识块到数据库失败: %w", err)
	}
	return nil
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

func ensureSessionWritable(ctx context.Context, db execQuerier, chatId string, userID int64) error {
	row := db.QueryRow(ctx, `SELECT user_id, is_active FROM chat_sessions WHERE session_id = $1 LIMIT 1`, chatId)

	var ownerID int64
	var isActive bool
	if err := row.Scan(&ownerID, &isActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if ownerID != userID {
		return ErrSessionAccessDenied
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

const (
	defaultSessionTitle = "新对话"
	defaultSessionMode  = sessionmode.DefaultKey
)

type scoredKnowledgeChunk struct {
	types.KnowledgeChunk
	Score float64
}

func (vs *VectorStore) fetchPublicKnowledge(ctx context.Context, queryEmbeddingVector pgvector.Vector, topK int, userID *int64) ([]scoredKnowledgeChunk, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sql := `SELECT id, title, content, embedding <-> $1 AS score
FROM knowledge_base
WHERE (user_id = 1`
	args := []any{queryEmbeddingVector}
	if userID != nil {
		sql += ` OR user_id = $2`
		args = append(args, *userID)
	}
	sql += `) AND embedding IS NOT NULL`
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

	// 定义插入 SQL 语句
	insertSQL := `INSERT INTO knowledge_base (title, content, embedding, user_id) VALUES ($1, $2, $3, $4)`
	// 遍历所有知识块
	for _, item := range embeddings {
		// 执行 SQL 插入语句
		if _, err := tx.Exec(ctx, insertSQL, title, item.chunk, item.embedding, userID); err != nil {
			return fmt.Errorf("保存知识块到数据库失败: %w", err)
		}
	}

	// 提交事务
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("提交事务失败: %w", commitErr)
	}

	return nil
}

func rollbackTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
