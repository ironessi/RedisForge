package v1

import "github.com/gogf/gf/v2/frame/g"

// ProfileReq 是获取当前登录用户信息的请求。
// 这个接口不需要前端传 userId，userId 会从 JWT 中间件解析出来。
type ProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"User" summary:"用户信息"`
}

type ProfileRes struct {
	UserId   uint64 `json:"userId"`   // 用户 ID
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"`
}

// UpdateProfileReq 是更新当前登录用户资料的请求。
// userId 不由前端传入，而是从 JWT 鉴权上下文中获取。
type UpdateProfileReq struct {
	g.Meta   `path:"/user/profile" method:"put" tags:"User" summary:"更新用户信息"`
	Nickname string `json:"nickname" v:"required|length:3,20#请输入昵称|昵称长度为3~20位"`
}

// UpdateProfileRes 是更新用户资料的响应。
type UpdateProfileRes struct{}
