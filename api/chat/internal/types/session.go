// Package types 提供 AI 面试官聊天相关的数据结构定义，
// 包括会话管理和存储接口。
package types

import (
	"github.com/sashabaranov/go-openai"
)

// ChatSession 聊天会话结构体
// V1 版本，支持“会话”，但无法实现“消息”，即无法持久对话，窗口刷新或服务器重启后，会话丢失
// 定义了单个聊天会话的消息历史记录（系统消息 + 用户消息 + AI回复）
// 这是多轮对话的核心数据结构，保存对话的完整上下文
// ChatSession 聊天会话结构体
// 定义了单个聊天会话的消息历史记录
// type ChatSession struct {
// 	Message []openai.ChatCompletionMessage `json:"message"`
// }

// // SessionStore 会话存储接口
// // 定义了会话数据的获取和保存操作，支持多轮对话功能
// type SessionStore interface {
// 	GetSession(chatId string) (*ChatSession, error)
// 	SaveSession(chatId string, session *ChatSession) error
// }

// VectorMessage V2 版本，引入向量数据库
// VectorMessage 向量消息结构体
type VectorMessage struct {
	Role    string `json:"role"`    // 角色：user、assistant
	Content string `json:"content"` // 内容
}

// KnowledgeChunk 新增：RAG 本地知识库会话结构体
// 主要是用来存储知识库内容到数据库中的 knowledge_base 表中
type KnowledgeChunk struct {
	ID      int64  `json:"id"`      // 知识块的ID(也就是数据库主键)
	Title   string `json:"title"`   // 知识块的标题（通常是文件名）
	Content string `json:"content"` // 知识块的内容
}

// VectorSession 向量会话存储接口
type VectorSession interface {
	GetSession(chatId string) ([]openai.ChatCompletionMessage, error)   // 获取历史信息
	SaveMessage(chatId, role, content string) error                     // 保存消息
	SaveKnowledge(title, content string) error                          // 保存知识到 RAG 本地知识库中
	RetrieveKnowledge(query string, topK int) ([]KnowledgeChunk, error) // 检索知识库
}
