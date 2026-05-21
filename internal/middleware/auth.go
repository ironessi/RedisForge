package middleware

import (
	"strings"

	authLogic "redis-demo/internal/logic/auth"
	jwtLogic "redis-demo/internal/logic/jwt"

	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	ContextUserId   = "userId"   // 用户 ID
	ContextUsername = "username" // 用户名
)

// Auth 是 JWT 鉴权中间件。
// 它会校验 Authorization: Bearer <token>，并把用户信息写入请求上下文。
func Auth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteExit(ghttp.DefaultHandlerResponse{ // 返回错误信息
			Code:    401,
			Message: "缺少 Authorization 请求头",
		})
		return
	}

	// Authorization 的标准格式是：Bearer token字符串
	parts := strings.SplitN(authHeader, " ", 2) // 按空格分割成两部分
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.WriteExit(ghttp.DefaultHandlerResponse{ // 返回错误信息
			Code:    401,
			Message: "Authorization 请求头格式错误",
		})
		return
	}

	// 解析 JWT，校验签名和过期时间
	claims, err := jwtLogic.ParseToken(r.Context(), parts[1]) // parts[1] 是 token 字符串
	if err != nil {
		r.Response.WriteExit(ghttp.DefaultHandlerResponse{ // 返回错误信息
			Code:    401,
			Message: "无效的 JWT token ",
		})
		return
	}

	// 检查 token 是否已经被加入 Redis 黑名单。
	// 如果用户已经退出登录，这个 token 即使还没过期，也不能继续使用。
	blacklisted, err := authLogic.IsTokenBlacklisted(r.Context(), parts[1]) // parts[1] 是 token 字符串
	if err != nil {
		r.Response.WriteJsonExit(ghttp.DefaultHandlerResponse{ // 返回错误信息
			Code:    500,
			Message: "token状态错误",
		})
		return
	}
	if blacklisted {
		r.Response.WriteJsonExit(ghttp.DefaultHandlerResponse{ // 返回错误信息
			Code:    401,
			Message: "token已退出登录",
		})
		return
	}

	// 将用户信息写入请求上下文，后续 controller 可以直接读取。
	r.SetCtxVar(ContextUserId, claims.UserId)     // 把用户 ID 写入请求上下文，方便后续处理函数使用
	r.SetCtxVar(ContextUsername, claims.Username) // 把用户名写入请求上下文，方便后续处理函数使用

	r.Middleware.Next() // 继续执行后续的处理函数
}
