// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Notification is the golang structure for table notification.
type Notification struct {
	Id            uint64      `json:"id"            orm:"id"              description:"通知ID"`         // 通知ID
	ReceiverId    uint64      `json:"receiverId"    orm:"receiver_id"     description:"接收人ID"`        // 接收人ID
	ActorId       uint64      `json:"actorId"       orm:"actor_id"        description:"触发人ID"`        // 触发人ID
	Type          string      `json:"type"          orm:"type"            description:"通知类型"`         // 通知类型
	Content       string      `json:"content"       orm:"content"         description:"通知内容"`         // 通知内容
	RelatedTaskId uint64      `json:"relatedTaskId" orm:"related_task_id" description:"关联任务ID"`       // 关联任务ID
	IsRead        uint        `json:"isRead"        orm:"is_read"         description:"是否已读：0未读/1已读"` // 是否已读：0未读/1已读
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"      description:"创建时间"`         // 创建时间
	ReadAt        *gtime.Time `json:"readAt"        orm:"read_at"         description:"已读时间"`         // 已读时间
}
