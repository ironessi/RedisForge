package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// 1. 定义创建任务每分钟最大次数
const taskCreateLimit = 10

// 2. 定义限流窗口 TTL
const rateLimitKeyTTL = 60 * time.Second

const loginLimit = 5

// 3. 生成创建任务限流 key
func taskCreateRateKey(userId uint64, minute int64) string {
	// 返回 rate:task:create:{userId}:{minute}
	return fmt.Sprintf("rate:task:create:%d:%d", userId, minute)
}

// CheckTaskCreate 检查当前用户是否还能创建任务。
// 规则：同一个用户每分钟最多创建 10 个任务。
func CheckTaskCreate(ctx context.Context, userId uint64) error {
	// 1. 计算当前分钟窗口
	minute := time.Now().Unix() / 60
	// 2. 生成 Redis key
	key := taskCreateRateKey(userId, minute)
	// 3. 对 key 执行 INCR
	count, err := g.Redis().Incr(ctx, key) // 3. 对 key 执行 INCR，获取当前计数
	if err != nil {
		return err
	}
	// 4. 如果是第一次计数，设置 60 秒过期时间
	if count == 1 {
		if _, err := g.Redis().Expire(ctx, key, int64(rateLimitKeyTTL.Seconds())); err != nil {
			return err
		}
	}
	// 5. 如果计数超过上限，返回“请求过于频繁，请稍后再试”
	if count > taskCreateLimit {
		return fmt.Errorf("请求过于频繁，请稍后再试")
	}

	// 6. 未超限，返回 nil
	return nil
}

// 2. 生成登录限流 key
func loginRateKey(ip string, minute int64) string {
	// 返回 rate:login:{ip}:{minute}
	return fmt.Sprintf("rate:login:%s:%d", ip, minute)
}

// CheckLogin 检查当前 IP 是否还能继续登录。
func CheckLogin(ctx context.Context, ip string) error {
	// 1. 计算当前分钟窗口
	minute := time.Now().Unix() / 60
	// 2. 生成 Redis key
	key := loginRateKey(ip, minute)
	// 3. 对 key 执行 INCR
	count, err := g.Redis().Incr(ctx, key)
	if err != nil {
		return err
	}

	// 4. 第一次计数时设置 60 秒过期时间
	if count == 1 {
		if _, err := g.Redis().Expire(ctx, key, int64(rateLimitKeyTTL.Seconds())); err != nil {
			return err
		}
	}
	// 5. 超过 5 次时返回“登录过于频繁，请稍后再试”
	if count > loginLimit {
		return gerror.New("登录过于频繁，请稍后再试")
	}

	// 6. 未超限返回 nil
	return nil
}
