package notification

import (
	"context"
	v1 "redis-demo/api/notification/v1"
	notificationLogic "redis-demo/internal/logic/notification"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Read 将当前用户的一条通知标记为已读。
func (c *ControllerV1) Read(ctx context.Context, req *v1.ReadReq) (res *v1.ReadRes, err error) {
	// 1. 从 JWT 上下文读取当前 userId
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	// 2. 未登录时返回“请先登录”
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}
	// 3. 调用 notificationLogic.MarkAsRead(ctx, userId, req.NotificationId)
	err = notificationLogic.MarkAsRead(ctx, userId, req.NotificationId)
	if err != nil {
		return nil, err
	}

	// 4. 成功后返回空响应
	return &v1.ReadRes{}, nil
}
