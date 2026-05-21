package auth

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const captchaExpire = 5 * time.Minute // 验证码过期时间，设置为 5 分钟，实际项目中可以根据需要调整

// GenerateCaptcha 生成登录验证码，并写入 Redis。
// 当前学习阶段会把验证码返回给前端，真实项目里应该通过短信/邮箱发送。
func GenerateCaptcha(ctx context.Context, username string) (string, error) {
	// 生成 100000 到 999999 之间的 6 位数字验证码。
	code := fmt.Sprintf("%06d", rand.IntN(1000000))
	key := captchaKey(username) // 生成 Redis key，例如 "captcha:username"

	// 将验证码写入 Redis，并设置过期时间，单位为秒。
	err := g.Redis().SetEX(ctx, key, code, int64(captchaExpire.Seconds()))
	if err != nil {
		return "", err
	}
	return code, nil
}

// VerifyCaptcha 校验验证码是否正确。
// 后面登录时可以调用这个函数，实现验证码登录或登录前置校验。
func VerifyCaptcha(ctx context.Context, username, code string) (bool, error) {
	key := captchaKey(username) // 生成 Redis key，例如 "captcha:username"

	// 从 Redis 获取验证码。
	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		return false, err
	}
	if value.IsNil() { // key 不存在，说明验证码过期了
		return false, nil

	}
	return value.String() == code, nil // 验证码匹配返回 true，否则返回 false
}

// captchaKey 统一生成验证码 Redis key。
// 例如：auth:captcha:test001
func captchaKey(username string) string {
	return fmt.Sprintf("auth:captcha:%s", username)
}
