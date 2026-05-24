package v1

import "github.com/gogf/gf/v2/frame/g"

// CreateReq 是创建团队请求。
// ownerId 不由前端传入，而是从 JWT 中间件解析出的当前用户中获取。
type CreateReq struct {
	g.Meta `path:"/teams" method:"post" tags:"Team" summary:"创建团队"`
	Name   string `json:"name" v:"required|length:3,20#请输入团队名称|团队名称长度为3~20位"`
}

// CreateRes 是创建团队返回。
type CreateRes struct {
	TeamId uint64 `json:"teamId"`
}

// AddMemberReq 是添加团队成员请求。
// teamId 来自路径参数，userId 是要添加进团队的用户ID。
type AddMemberReq struct {
	g.Meta `path:"/teams/{teamId}/members" method:"post" tags:"Team" summary:"添加团队成员"`
	TeamId uint64 `json:"teamId" v:"required|min:1#团队ID不能为空|团队ID不合法"`
	UserId uint64 `json:"userId" v:"required|min:1#用户ID不能为空|用户ID不合法"`
}

// AddMemberRes 是添加团队成员响应。
type AddMemberRes struct{}

// MemberItem 是团队成员列表项。
type MemberItem struct {
	UserId   uint64 `json:"userId"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"` // owner 或 member
}

// MembersReq 是查询团队成员请求。
// teamId 来自路径参数。
type MembersReq struct {
	g.Meta `path:"/teams/{teamId}/members" method:"get" tags:"Team" summary:"查询团队成员"`
	TeamId uint64 `json:"teamId" v:"required|min:1#团队ID不能为空|团队ID不合法"`
}

// MembersRes 是查询团队成员响应。
type MembersRes struct {
	Members []MemberItem `json:"members"` // 切片吗？是的，团队成员列表可能有多个成员，所以用切片来表示。
}
