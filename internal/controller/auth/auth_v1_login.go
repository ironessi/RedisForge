package auth

import (
	"context"
	v1 "redis-demo/api/auth/v1"
	"redis-demo/internal/logic/auth"
	"redis-demo/internal/logic/ratelimit"

	"github.com/gogf/gf/v2/frame/g"
)

// Login 处理用户登录请求。
// 当前先只打通接口结构，后面再实现密码校验和 JWT 生成。
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	//登陆限流
	ip := g.RequestFromCtx(ctx).GetClientIp() // 获取客户端 IP 地址
	if err := ratelimit.CheckLogin(ctx, ip); err != nil {
		return nil, err
	}

	token, err := auth.Login(ctx, req.Username, req.Password, req.Captcha) // 传入验证码参数
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{
		Token: token,
	}, nil
}
