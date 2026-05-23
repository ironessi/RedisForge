// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// TeamMember is the golang structure for table team_member.
type TeamMember struct {
	Id        uint64      `json:"id"        orm:"id"         description:"主键ID"`            // 主键ID
	TeamId    uint64      `json:"teamId"    orm:"team_id"    description:"团队ID"`            // 团队ID
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"用户ID"`            // 用户ID
	Role      string      `json:"role"      orm:"role"       description:"角色：owner/member"` // 角色：owner/member
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"加入时间"`            // 加入时间
}
