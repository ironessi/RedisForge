package auth

import (
	"context"
	v1 "redis-demo/api/auth/v1"
	"redis-demo/internal/logic/auth"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Logout 处理用户退出登录。
// 它会将当前请求携带的 JWT 加入 Redis 黑名单，使该 token 后续无法继续访问接口。
func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	// 从请求上下文中获取当前 token，通常是从 Authorization 请求头解析出来的。
	authHeader := g.RequestFromCtx(ctx).Header.Get("Authorization")
	if authHeader == "" {
		return nil, gerror.New("缺少 Authorization 请求头")
	}
	// Authorization 的标准格式是：Bearer token字符串。
	parts := strings.SplitN(authHeader, " ", 2) // 按空格分割成两部分，第一部分是 "Bearer"，第二部分是 token 字符串。
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, gerror.New("Authorization 请求头格式错误，应该是 Bearer token字符串")
	}

	// 将当前 token 写入 Redis 黑名单。
	if err := auth.AddTokenToBlacklist(ctx, parts[1]); err != nil {
		return nil, err
	}
	return &v1.LogoutRes{}, nil
}
