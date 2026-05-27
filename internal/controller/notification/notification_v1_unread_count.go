package notification

import (
	"context"
	v1 "redis-demo/api/notification/v1"
	notificationLogic "redis-demo/internal/logic/notification"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// UnreadCount 查询当前用户的未读通知数量。
func (c *ControllerV1) UnreadCount(ctx context.Context, req *v1.UnreadCountReq) (res *v1.UnreadCountRes, err error) {
	// 1. 从 JWT 鉴权上下文读取当前 userId
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	// 2. 如果 userId 为 0，返回“请先登录”
	if userId == 0 {
		return nil, gerror.New("请先登录")

	}

	// 3. 调用 notificationLogic.GetUnreadCount 查询未读数量
	count, err := notificationLogic.GetUnreadCount(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 4. 将未读数量封装为 UnreadCountRes 返回
	return &v1.UnreadCountRes{
		Count: count,
	}, nil
}
