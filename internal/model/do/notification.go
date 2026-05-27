// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Notification is the golang structure of table notification for DAO operations like Where/Data.
type Notification struct {
	g.Meta        `orm:"table:notification, do:true"`
	Id            any         // 通知ID
	ReceiverId    any         // 接收人ID
	ActorId       any         // 触发人ID
	Type          any         // 通知类型
	Content       any         // 通知内容
	RelatedTaskId any         // 关联任务ID
	IsRead        any         // 是否已读：0未读/1已读
	CreatedAt     *gtime.Time // 创建时间
	ReadAt        *gtime.Time // 已读时间
}
