// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TeamDao is the data access object for the table team.
type TeamDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TeamColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TeamColumns defines and stores column names for the table team.
type TeamColumns struct {
	Id        string // 团队ID
	Name      string // 团队名称
	OwnerId   string // 创建者用户ID
	CreatedAt string // 创建时间
	UpdatedAt string // 更新时间
}

// teamColumns holds the columns for the table team.
var teamColumns = TeamColumns{
	Id:        "id",
	Name:      "name",
	OwnerId:   "owner_id",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

// NewTeamDao creates and returns a new DAO object for table data access.
func NewTeamDao(handlers ...gdb.ModelHandler) *TeamDao {
	return &TeamDao{
		group:    "default",
		table:    "team",
		columns:  teamColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TeamDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TeamDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TeamDao) Columns() TeamColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TeamDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TeamDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TeamDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
