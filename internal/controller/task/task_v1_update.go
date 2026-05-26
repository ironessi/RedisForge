package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Update 处理任务基本信息编辑请求，状态流转由 UpdateStatus 单独处理。
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	operatorId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if operatorId == 0 {
		return nil, gerror.New("请先登录")
	}

	err = taskLogic.UpdateTask(
		ctx,
		operatorId,
		req.TaskId,
		req.Title,
		req.Description,
		req.AssigneeId,
		req.Priority,
	)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateRes{}, nil
}
