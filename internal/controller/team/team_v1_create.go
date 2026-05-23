package team

import (
	"context"
	v1 "redis-demo/api/team/v1"
	teamLogic "redis-demo/internal/logic/team"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Create 处理创建团队请求。
// 当前登录用户会自动成为团队 owner，并写入团队成员 Redis Set
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	// 从 JWT 鉴权上下文中取出当前用户 ID。
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 创建团队，并将当前用户作为 owner 加入团队成员。
	teamId, err := teamLogic.CreateTeam(ctx, userId, req.Name)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{
		TeamId: teamId,
	}, nil

}
