package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	// 1. 从 JWT 上下文读取 userId
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()

	// 2. userId 为 0 时返回“请先登录”
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}
	// 3. 调用 taskLogic.GetTasks
	tasks, err := taskLogic.GetTasks(ctx, userId, req.TeamId)
	if err != nil {
		return nil, err
	}
	// 4. 返回 &v1.ListRes{Tasks: tasks}
	return &v1.ListRes{
		Tasks: tasks,
	}, nil
}
