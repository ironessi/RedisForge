// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Id           uint64      `json:"id"           orm:"id"            description:"用户ID"`       // 用户ID
	Username     string      `json:"username"     orm:"username"      description:"用户名"`        // 用户名
	PasswordHash string      `json:"passwordHash" orm:"password_hash" description:"密码哈希"`       // 密码哈希
	Nickname     string      `json:"nickname"     orm:"nickname"      description:"昵称"`         // 昵称
	Status       int         `json:"status"       orm:"status"        description:"状态：1正常，0禁用"` // 状态：1正常，0禁用
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:"创建时间"`       // 创建时间
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:"更新时间"`       // 更新时间
}
