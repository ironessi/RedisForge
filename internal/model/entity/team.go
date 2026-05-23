// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Team is the golang structure for table team.
type Team struct {
	Id        uint64      `json:"id"        orm:"id"         description:"团队ID"`    // 团队ID
	Name      string      `json:"name"      orm:"name"       description:"团队名称"`    // 团队名称
	OwnerId   uint64      `json:"ownerId"   orm:"owner_id"   description:"创建者用户ID"` // 创建者用户ID
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"创建时间"`    // 创建时间
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"更新时间"`    // 更新时间
}
