package task

import (
	"context"
	v1 "redis-demo/api/task/v1"
	taskLogic "redis-demo/internal/logic/task"
	"redis-demo/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Hot 查询指定团队的热门任务排行榜。
// 当前登录用户必须属于该团队，排行榜数据来自 Redis Sorted Set。
func (c *ControllerV1) Hot(ctx context.Context, req *v1.HotReq) (res *v1.HotRes, err error) {
	// 从 JWT 鉴权上下文读取当前用户 ID，避免未登录用户查看团队数据。
	userId := g.RequestFromCtx(ctx).GetCtxVar(middleware.ContextUserId).Uint64()
	if userId == 0 {
		return nil, gerror.New("请先登录")
	}

	// 调用业务逻辑：校验团队权限，并查询 Redis 中的热门任务排行。
	tasks, err := taskLogic.GetHotTasks(ctx, userId, req.TeamId)
	if err != nil {
		return nil, err
	}

	// 将热门任务列表包装成接口响应返回给前端。
	return &v1.HotRes{
		Tasks: tasks,
	}, nil
}
