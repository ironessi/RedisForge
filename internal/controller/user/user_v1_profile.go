package user

import (
	"context"
	v1 "redis-demo/api/user/v1"
	"redis-demo/internal/logic/user"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Profile 获取当前登录用户信息。
// userId 来自 JWT 鉴权中间件，而不是前端传参，避免越权访问。
func (c *Controller) Profile(ctx context.Context, req *v1.ProfileReq) (*v1.ProfileRes, error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64() // 从请求上下文中获取用户 ID
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	return user.GetProfile(ctx, userId)
}
