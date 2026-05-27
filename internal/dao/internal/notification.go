// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// NotificationDao is the data access object for the table notification.
type NotificationDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  NotificationColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// NotificationColumns defines and stores column names for the table notification.
type NotificationColumns struct {
	Id            string // 通知ID
	ReceiverId    string // 接收人ID
	ActorId       string // 触发人ID
	Type          string // 通知类型
	Content       string // 通知内容
	RelatedTaskId string // 关联任务ID
	IsRead        string // 是否已读：0未读/1已读
	CreatedAt     string // 创建时间
	ReadAt        string // 已读时间
}

// notificationColumns holds the columns for the table notification.
var notificationColumns = NotificationColumns{
	Id:            "id",
	ReceiverId:    "receiver_id",
	ActorId:       "actor_id",
	Type:          "type",
	Content:       "content",
	RelatedTaskId: "related_task_id",
	IsRead:        "is_read",
	CreatedAt:     "created_at",
	ReadAt:        "read_at",
}

// NewNotificationDao creates and returns a new DAO object for table data access.
func NewNotificationDao(handlers ...gdb.ModelHandler) *NotificationDao {
	return &NotificationDao{
		group:    "default",
		table:    "notification",
		columns:  notificationColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *NotificationDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *NotificationDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *NotificationDao) Columns() NotificationColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *NotificationDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *NotificationDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *NotificationDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
