package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	"redis-demo/internal/logic/ratelimit"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Create 处理创建任务请求。
// 当前登录用户必须属于指定团队；创建成功后会写入任务记录和团队动态。
// 进入创建逻辑前，会先按用户维度做 Redis 限流，限制同一用户每分钟最多创建 10 个任务。
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	creatorId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if creatorId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 创建任务是写接口，先限流，避免短时间内大量写 MySQL 和动态流。
	if err := ratelimit.CheckTaskCreate(ctx, creatorId); err != nil {
		return nil, err
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
