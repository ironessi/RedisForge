package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Create 处理创建任务请求。
// 当前登录用户必须属于指定团队；创建成功后会写入任务记录和团队动态。
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	creatorId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if creatorId == 0 {
		return nil, gerror.New("请先登录")
	}

	taskId, err := taskLogic.CreateTask(
		ctx,
		creatorId,
		req.TeamId,
		req.Title,
		req.Description,
		req.AssigneeId,
		req.Priority,
	)
	if err != nil {
		return nil, err
	}

	return &v1.CreateRes{
		TaskId: taskId,
	}, nil
}
