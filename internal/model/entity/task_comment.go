// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskComment is the golang structure for table task_comment.
type TaskComment struct {
	Id        uint64      `json:"id"        orm:"id"         description:"评论ID"`   // 评论ID
	TaskId    uint64      `json:"taskId"    orm:"task_id"    description:"任务ID"`   // 任务ID
	TeamId    uint64      `json:"teamId"    orm:"team_id"    description:"团队ID"`   // 团队ID
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"评论用户ID"` // 评论用户ID
	Content   string      `json:"content"   orm:"content"    description:"评论内容"`   // 评论内容
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`   // 创建时间
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`   // 更新时间
}
