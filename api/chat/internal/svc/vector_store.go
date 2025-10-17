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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"

	"GoZero-AI/api/chat/internal/config"
	"GoZero-AI/api/chat/internal/types"
)

// VectorStore 向量存储结构
type VectorStore struct {
	Pool           *pgxpool.Pool  // 数据库连接池
	OpenAIClient   *openai.Client // OpenAI客户端
	EmbeddingModel string         // 向量模型名称
}

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

// generateEmbedding 清洗 content 文本，并将其转换为 []byte 类型的向量
func (vs *VectorStore) generateEmbedding(content string) ([]byte, error) {
	if content == "" {
		return nil, nil
	}

	// 使用 OpenAI 客户端生成向量, 返回一个 EmbeddingResponse 结构体
	// 这个结构体包含了一个 Data 字段，其中 Data 字段是一个 Embedding 结构体的切片
	// 每个 Embedding 结构体包含一个 Embedding 字段，这个字段是一个 []float32 类型的向量
	embedding, err := vs.OpenAIClient.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Model: openai.EmbeddingModel(vs.EmbeddingModel), // 使用配置文件中的模型名称
			Input: []string{content},                        // 用户对话的上下文内容
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API 错误，生成向量失败: %w", err)
	}

	if len(embedding.Data) == 0 {
		return nil, errors.New("OpenAI API 错误，生成向量失败: 没有返回数据")
	}

	// 将向量转换为 JSON 格式
	embeddingJSON, err := json.Marshal(embedding.Data[0].Embedding)
	if err != nil {
		return nil, fmt.Errorf("序列化嵌入失败: %w", err)
	}

	// 返回向量数据
	return embeddingJSON, nil
}

// SaveMessage 保存消息到向量数据库
// 新增 RAG 本地知识库后，进行了升级
func (vs *VectorStore) SaveMessage(chatId, role, content string) error {
	// 步骤1：生成文本向量
	embedding, err := vs.generateEmbedding(content)
	if err != nil {
		return fmt.Errorf("生成嵌入失败: %w", err)
	}

	// 步骤2：向量序列化
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("序列化嵌入失败: %w", err)
	}

	// 步骤3：数据库存储
	sql := `INSERT INTO vector_store (chat_id, role, content, embedding, source_type) 
            VALUES ($1, $2, $3, $4, 'message')`
	_, err = vs.Pool.Exec(context.Background(), sql,
		chatId, role, content, embeddingJSON)

	return err
}

// GetMessage 获取消息从向量数据库
func (vs *VectorStore) GetMessage(chatId string, limit int) ([]types.VectorMessage, error) {
	// 1. 构建一条 SQL 语句，从表中获取消息
	sql := `SELECT role, content FROM vector_store WHERE chat_id = $1 ORDER BY created_at DESC LIMIT $2`

	// 2. 传入指定用户 ID，获取他的历史对话数据，并做出限制，我们肯定不希望数据太多导致准确性降低
	rows, err := vs.Pool.Query(context.Background(), sql, chatId, limit)
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
	// 1. 将知识块内容转换为向量表示
	// generateEmbedding()返回已序列化的JSON向量数据
	embeddingJSON, err := vs.generateEmbedding(chunk)
	if err != nil {
		return fmt.Errorf("知识块向量化失败: %w", err)
	}

	// 2. 将知识块和向量数据存储到knowledge_base表
	// 与vector_store表分离，专门存储知识库数据
	sql := `INSERT INTO knowledge_base (title, content, embedding) VALUES ($1, $2, $3)`
	_, err = vs.Pool.Exec(context.Background(), sql, title, chunk, embeddingJSON)
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
	// 1. 将用户查询转换为向量表示
	// 使用相同的embedding模型确保向量空间一致性
	queryEmbeddingJSON, err := vs.generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("用户查询向量化失败: %w", err)
	}

	// 2. 构建SQL查询语句，使用向量相似度排序
	// <->是PostgreSQL的向量距离操作符，计算余弦距离
	// 距离越小表示相似度越高，ORDER BY升序排列
	sql := `SELECT id, title, content 
          FROM knowledge_base 
          ORDER BY embedding::jsonb::text <-> $1::text
          LIMIT $2`

	// 3. 执行数据库查询
	rows, err := vs.Pool.Query(context.Background(), sql, queryEmbeddingJSON, topK)
	if err != nil {
		return nil, fmt.Errorf("知识库检索失败: %w", err)
	}
	defer rows.Close()

	// 4. 遍历查询结果，构建知识块列表
	var results []types.KnowledgeChunk
	for rows.Next() {
		var id int64
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			return nil, fmt.Errorf("扫描知识块数据失败: %w", err)
		}
		// 构建知识块结构体，包含ID、标题和内容
		results = append(results, types.KnowledgeChunk{
			ID:      id,      // 知识块唯一标识
			Title:   title,   // 文档标题，用于结果展示
			Content: content, // 知识块内容，用于注入AI上下文
		})
	}

	return results, nil
}
