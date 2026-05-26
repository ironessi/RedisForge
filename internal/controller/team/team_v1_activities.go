package team

import (
	"context"
	v1 "redis-demo/api/team/v1"
	teamLogic "redis-demo/internal/logic/team"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 当前登录用户必须属于该团队，才可以查看动态。
func (c *ControllerV1) Activities(ctx context.Context, req *v1.ActivitiesReq) (res *v1.ActivitiesRes, err error) {
	// 从 JWT 鉴权上下文读取当前用户 ID，确认请求来自登录用户。
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 从 Redis List 查询该团队最近动态。
	activities, err := teamLogic.GetActivities(ctx, userId, req.TeamId)
	if err != nil {
		return nil, err
	}
	return &v1.ActivitiesRes{
		Activities: activities,
	}, nil
}
