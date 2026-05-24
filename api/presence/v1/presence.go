package v1

import "github.com/gogf/gf/v2/frame/g"

// HeartbeatReq 是用户在线心跳请求。
// teamId 表示用户当前所在团队页面或工作区。
type HeartbeatReq struct {
	g.Meta `path:"/presence/heartbeat" method:"post" tags:"Presence" summary:"用户在线心跳"`
	TeamId uint64 `json:"teamId" v:"required|min:1#请选择团队|团队ID不合法"`
}

// HeartbeatRes 是用户在线心跳响应。
type HeartbeatRes struct{}

// OnlineMemberItem 是在线成员列表项。
type OnlineMembersItem struct {
	UserId   uint64 `json:"userId"`   // 用户ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 其他用户信息字段...
}

// OnlineMembersReq 是查询团队在线成员请求。
type OnlineMembersReq struct {
	g.Meta `path:"/teams/{teamId}/online-members" method:"get" tags:"Presence" summary:"查询团队在线成员"`
	TeamId uint64 `json:"teamId" v:"required|min:1#请选择团队|团队ID不合法"`
}

// OnlineMembersRes 是查询团队在线成员响应。
type OnlineMembersRes struct {
	Members []OnlineMembersItem `json:"members"` // 在线成员列表
}
