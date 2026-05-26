// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Task is the golang structure for table task.
type Task struct {
	Id          uint64      `json:"id"          orm:"id"          description:"任务ID"`               // 任务ID
	TeamId      uint64      `json:"teamId"      orm:"team_id"     description:"所属团队ID"`             // 所属团队ID
	Title       string      `json:"title"       orm:"title"       description:"任务标题"`               // 任务标题
	Description string      `json:"description" orm:"description" description:"任务描述"`               // 任务描述
	CreatorId   uint64      `json:"creatorId"   orm:"creator_id"  description:"创建人ID"`              // 创建人ID
	AssigneeId  uint64      `json:"assigneeId"  orm:"assignee_id" description:"负责人ID"`              // 负责人ID
	Status      string      `json:"status"      orm:"status"      description:"状态：todo/doing/done"` // 状态：todo/doing/done
	Priority    int         `json:"priority"    orm:"priority"    description:"优先级：1低，2中，3高"`       // 优先级：1低，2中，3高
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  description:"创建时间"`               // 创建时间
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  description:"更新时间"`               // 更新时间
}
