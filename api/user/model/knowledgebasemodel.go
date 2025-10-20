package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ KnowledgeBaseModel = (*customKnowledgeBaseModel)(nil)

type (
	// KnowledgeBaseModel is an interface to be customized, add more methods here,
	// and implement the added methods in customKnowledgeBaseModel.
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
