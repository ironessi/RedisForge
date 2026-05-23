package user

import (
	"context"
	v1 "redis-demo/api/user/v1"
	userLogic "redis-demo/internal/logic/user"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// UpdateProfile 更新当前登录用户资料。
// userId 从 JWT 鉴权上下文中读取，不允许前端传 userId，避免越权更新。
func (c *Controller) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (*v1.UpdateProfileRes, error) {
	// 从请求上下文中取出 JWT 中间件写入的 userId。
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}
	// 更新 MySQL，并删除 Redis 中的用户资料缓存。
	if err := userLogic.UpdateProfile(ctx, userId, req.Nickname); err != nil {
		return nil, err
	}
	return &v1.UpdateProfileRes{}, nil
}
