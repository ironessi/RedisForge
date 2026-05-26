// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Task is the golang structure of table task for DAO operations like Where/Data.
type Task struct {
	g.Meta      `orm:"table:task, do:true"`
	Id          any         // 任务ID
	TeamId      any         // 所属团队ID
	Title       any         // 任务标题
	Description any         // 任务描述
	CreatorId   any         // 创建人ID
	AssigneeId  any         // 负责人ID
	Status      any         // 状态：todo/doing/done
	Priority    any         // 优先级：1低，2中，3高
	CreatedAt   *gtime.Time // 创建时间
	UpdatedAt   *gtime.Time // 更新时间
}
