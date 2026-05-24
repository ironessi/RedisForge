package team

import (
	"context"
	v1 "redis-demo/api/team/v1"
	teamLogic "redis-demo/internal/logic/team"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Members 查询团队成员列表。
// 当前只要求登录；后面可以进一步限制为团队成员才能查看。
func (c *ControllerV1) Members(ctx context.Context, req *v1.MembersReq) (res *v1.MembersRes, err error) {
	// 从 JWT 鉴权上下文中取出当前用户 ID，确保请求来自已登录用户。
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	members, err := teamLogic.GetMembers(ctx, req.TeamId)
	if err != nil {
		return nil, err
	}

	return &v1.MembersRes{
		Members: members,
	}, nil
}
