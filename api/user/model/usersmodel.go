package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	//
	// 注意：
	// `db/user.sql` 是 users 表的仓库事实源。若 live 数据库额外出现 email / updated_at
	// 之类 repo 未声明字段，应先按环境漂移核对，而不是直接把本地库结构回写到 repo。
	UsersModel interface {
		usersModel
		withSession(session sqlx.Session) UsersModel
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn),
	}
}

func (m *customUsersModel) withSession(session sqlx.Session) UsersModel {
	return NewUsersModel(sqlx.NewSqlConnFromSession(session))
}
