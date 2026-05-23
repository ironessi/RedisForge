// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// TeamMember is the golang structure of table team_member for DAO operations like Where/Data.
type TeamMember struct {
	g.Meta    `orm:"table:team_member, do:true"`
	Id        any         // 主键ID
	TeamId    any         // 团队ID
	UserId    any         // 用户ID
	Role      any         // 角色：owner/member
	CreatedAt *gtime.Time // 加入时间
}
