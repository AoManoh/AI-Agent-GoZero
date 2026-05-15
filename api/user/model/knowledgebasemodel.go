package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ KnowledgeBaseModel = (*customKnowledgeBaseModel)(nil)

type (
	// KnowledgeBaseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeBaseModel.
	//
	// 注意：
	// 当前 live knowledge_base.embedding 已升级为 vector(1536)，而 goctl 1.8.5
	// 还不能直接从 PostgreSQL vector 列重生 *_gen.go。除非先确认 model 快照已对齐，
	// 否则不要把这个 legacy model 接回运行时链路。
	KnowledgeBaseModel interface {
		knowledgeBaseModel
		withSession(session sqlx.Session) KnowledgeBaseModel
	}

	customKnowledgeBaseModel struct {
		*defaultKnowledgeBaseModel
	}
)

// NewKnowledgeBaseModel returns a model for the database table.
func NewKnowledgeBaseModel(conn sqlx.SqlConn) KnowledgeBaseModel {
	return &customKnowledgeBaseModel{
		defaultKnowledgeBaseModel: newKnowledgeBaseModel(conn),
	}
}

func (m *customKnowledgeBaseModel) withSession(session sqlx.Session) KnowledgeBaseModel {
	return NewKnowledgeBaseModel(sqlx.NewSqlConnFromSession(session))
}
