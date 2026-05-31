// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TaskComment is the golang structure of table task_comment for DAO operations like Where/Data.
type TaskComment struct {
	g.Meta    `orm:"table:task_comment, do:true"`
	Id        any         // 评论ID
	TaskId    any         // 任务ID
	TeamId    any         // 团队ID
	UserId    any         // 评论用户ID
	Content   any         // 评论内容
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
}
