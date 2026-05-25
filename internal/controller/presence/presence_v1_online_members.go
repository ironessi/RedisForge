package presence

import (
	"context"
	v1 "redis-demo/api/presence/v1"
	presenceLogic "redis-demo/internal/logic/presence"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// OnlineMembers 查询团队在线成员。
// 当前要求已登录，后续可以限制只有团队成员可查看。
func (c *ControllerV1) OnlineMembers(ctx context.Context, req *v1.OnlineMembersReq) (res *v1.OnlineMembersRes, err error) {
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	users, err := presenceLogic.GetOnlineMembers(ctx, req.TeamId)
	if err != nil {
		return nil, err
	}

	// 预留容量但不提前创建元素，避免响应中出现空用户记录。
	members := make([]v1.OnlineMembersItem, 0, len(users))
	for _, user := range users {
		members = append(members, v1.OnlineMembersItem{
			UserId:   user.Id,
			Username: user.Username,
			Nickname: user.Nickname,
		})
	}

	return &v1.OnlineMembersRes{
		Members: members,
	}, nil
}
