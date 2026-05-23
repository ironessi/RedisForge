// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Team is the golang structure of table team for DAO operations like Where/Data.
type Team struct {
	g.Meta    `orm:"table:team, do:true"`
	Id        any         // 团队ID
	Name      any         // 团队名称
	OwnerId   any         // 创建者用户ID
	CreatedAt *gtime.Time // 创建时间
	UpdatedAt *gtime.Time // 更新时间
}
