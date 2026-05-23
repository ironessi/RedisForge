// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeamMemberDao is the data access object for the table team_member.
type TeamMemberDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TeamMemberColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TeamMemberColumns defines and stores column names for the table team_member.
type TeamMemberColumns struct {
	Id        string // 主键ID
	TeamId    string // 团队ID
	UserId    string // 用户ID
	Role      string // 角色：owner/member
	CreatedAt string // 加入时间
}

// teamMemberColumns holds the columns for the table team_member.
var teamMemberColumns = TeamMemberColumns{
	Id:        "id",
	TeamId:    "team_id",
	UserId:    "user_id",
	Role:      "role",
	CreatedAt: "created_at",
}

// NewTeamMemberDao creates and returns a new DAO object for table data access.
func NewTeamMemberDao(handlers ...gdb.ModelHandler) *TeamMemberDao {
	return &TeamMemberDao{
		group:    "default",
		table:    "team_member",
		columns:  teamMemberColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TeamMemberDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TeamMemberDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TeamMemberDao) Columns() TeamMemberColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TeamMemberDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TeamMemberDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *TeamMemberDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
