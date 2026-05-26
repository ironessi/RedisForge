// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TaskDao is the data access object for the table task.
type TaskDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  TaskColumns        // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// TaskColumns defines and stores column names for the table task.
type TaskColumns struct {
	Id          string // 任务ID
	TeamId      string // 所属团队ID
	Title       string // 任务标题
	Description string // 任务描述
	CreatorId   string // 创建人ID
	AssigneeId  string // 负责人ID
	Status      string // 状态：todo/doing/done
	Priority    string // 优先级：1低，2中，3高
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
}

// taskColumns holds the columns for the table task.
var taskColumns = TaskColumns{
	Id:          "id",
	TeamId:      "team_id",
	Title:       "title",
	Description: "description",
	CreatorId:   "creator_id",
	AssigneeId:  "assignee_id",
	Status:      "status",
	Priority:    "priority",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewTaskDao creates and returns a new DAO object for table data access.
func NewTaskDao(handlers ...gdb.ModelHandler) *TaskDao {
	return &TaskDao{
		group:    "default",
		table:    "task",
		columns:  taskColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *TaskDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *TaskDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *TaskDao) Columns() TaskColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *TaskDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *TaskDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *TaskDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
