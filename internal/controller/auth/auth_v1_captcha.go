package auth

import (
	"context"
	v1 "redis-demo/api/auth/v1"
	authLogic "redis-demo/internal/logic/auth" // 引入 auth 业务逻辑包
)

// Captcha 生成登录验证码。
// 当前学习阶段直接返回验证码，真实项目中应通过短信或邮箱发送。
func (c *ControllerV1) Captcha(ctx context.Context, req *v1.CaptchaReq) (res *v1.CaptchaRes, err error) {
	// 调用业务逻辑生成验证码，并写入 Redis 设置 TTL。
	code, err := authLogic.GenerateCaptcha(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	return &v1.CaptchaRes{
		Code: code,
	}, nil

}
