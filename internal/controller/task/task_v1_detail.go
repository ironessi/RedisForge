package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	// 1. 从 JWT 上下文读取当前 userId
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	// 2. 未登录则返回“请先登录”
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 3. 调用 taskLogic.GetTask(ctx, userId, req.TaskId)
	task, err := taskLogic.GetTask(ctx, userId, req.TaskId)
	if err != nil {
		return nil, err
	}
	// 4. 返回 &v1.DetailRes{Task: *task}
	return &v1.DetailRes{Task: *task}, nil
}
