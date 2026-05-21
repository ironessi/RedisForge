package auth

import (
	"context"
	v1 "redis-demo/api/auth/v1"
	"redis-demo/internal/logic/auth"
)

// Register 处理用户注册请求。
// 当前先只打通接口结构，真正的注册逻辑下一步再接入。
func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	userId, err := auth.Register(ctx, req.Username, req.Password, req.Nickname)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterRes{
		UserId: userId,
	}, nil
}
