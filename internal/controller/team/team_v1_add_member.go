package team

import (
	"context"
	v1 "redis-demo/api/team/v1"
	teamLogic "redis-demo/internal/logic/team"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// AddMember 处理添加团队成员请求。
// operatorId 来自 JWT，只有团队 owner 才能添加成员。
func (c *ControllerV1) AddMember(ctx context.Context, req *v1.AddMemberReq) (res *v1.AddMemberRes, err error) {
	// 从 JWT 鉴权上下文中取出当前操作用户 ID。
	operatorId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if operatorId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 添加目标用户到团队。
	if err := teamLogic.AddMember(ctx, operatorId, req.TeamId, req.UserId); err != nil {
		return nil, err
	}

	return &v1.AddMemberRes{}, nil
}
