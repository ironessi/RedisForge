package v1

import "github.com/gogf/gf/v2/frame/g"

type RegisterReq struct {
	g.Meta   `path:"/auth/register" method:"post" tags:"Auth" summary:"用户注册"`
	Username string `json:"username" v:"required|length:3,20#请输入用户名|用户名长度为3~20位"`
	Password string `json:"password" v:"required|length:6,20#请输入密码|密码长度为6~20位"`
	Nickname string `json:"nickname" v:"required|length:3,20#请输入昵称|昵称长度为3~20位"`
}

// RegisterRes 是用户注册响应。
type RegisterRes struct {
	UserId uint64 `json:"userId"`
}

type LoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"Auth" summary:"用户登录"`
	Username string `json:"username" v:"required|length:3,20#请输入用户名|用户名长度为3~20位"`
	Password string `json:"password" v:"required|length:6,20#请输入密码|密码长度为6~20位"`
	Captcha  string `json:"captcha" v:"required|length:6#请输入验证码|验证码长度为6位"`
}

// LoginRes 是用户登录响应。
// Token 是后续访问需要鉴权接口时携带的 JWT。
type LoginRes struct {
	Token string `json:"token"`
}

// LogoutReq 是用户退出登录请求。
// 退出登录需要携带 Authorization 请求头，服务端会将当前 token 加入 Redis 黑名单。
type LogoutReq struct {
	g.Meta `path:"/auth/logout" method:"post" tags:"Auth" summary:"用户退出登录"`
}

// LogoutRes 是用户退出登录响应。
type LogoutRes struct{}

// CaptchaReq 是获取登录验证码请求。
// 当前阶段用 username 作为验证码归属，后续可以替换成手机号或邮箱。
type CaptchaReq struct {
	g.Meta   `path:"/auth/captcha" method:"post" tags:"Auth" summary:"获取登录验证码"`
	Username string `json:"username" v:"required|length:3,20#请输入用户名|用户名长度为3~20位"`
}

// CaptchaRes 是获取验证码响应。
// 实际项目中不应该返回验证码，这里为了学习和测试方便才返回。
type CaptchaRes struct {
	Code string `json:"code"`
}
