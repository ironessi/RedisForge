package notification

import (
	"context"
	v1 "redis-demo/api/notification/v1"
	notificationLogic "redis-demo/internal/logic/notification"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// List 查询当前用户的通知列表。
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	// 1. 从 JWT 鉴权上下文读取当前 userId
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	// 2. 如果 userId 为 0，返回“请先登录”
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}
	// 3. 调用 notificationLogic.GetNotifications(ctx, userId)
	notifications, err := notificationLogic.GetNotifications(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 4. 将通知列表封装为 ListRes 返回
	return &v1.ListRes{
		Notifications: notifications,
	}, nil
}
