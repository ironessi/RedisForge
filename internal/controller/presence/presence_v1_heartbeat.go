package presence

import (
	"context"

	v1 "redis-demo/api/presence/v1"
	presenceLogic "redis-demo/internal/logic/presence"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Heartbeat 处理在线心跳。
// userId 来自 JWT，teamId 来自请求体。
func (c *ControllerV1) Heartbeat(ctx context.Context, req *v1.HeartbeatReq) (res *v1.HeartbeatRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	if err := presenceLogic.Heartbeat(ctx, userId, req.TeamId); err != nil {
		return nil, err
	}

	return &v1.HeartbeatRes{}, nil
}
