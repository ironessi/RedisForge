package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const tokenBlacklistExpire = 2 * time.Hour // 黑名单过期时间，设置为 JWT 的过期时间的两倍，确保 JWT 过期后黑名单也过期

// AddTokenToBlacklist 将 token 加入 Redis 黑名单。
// 黑名单 key 设置过期时间，避免 Redis 永久保存已经过期的 token。
func AddTokenToBlacklist(ctx context.Context, token string) error {
	key := tokenBlacklistKey(token) // 黑名单 key
	// value 只需要表示存在即可，这里存 "1"。
	err := g.Redis().SetEX(ctx, key, "1", int64(tokenBlacklistExpire.Seconds())) // 设置过期时间，单位为秒
	return err
}

// IsTokenBlacklisted 判断 token 是否在 Redis 黑名单中。
// 如果 key 存在，说明该 token 已退出登录，不允许继续访问。
func IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := tokenBlacklistKey(token)           // 黑名单 key，可以加前缀区分，例如 "blacklist:token:<token>"
	exists, err := g.Redis().Exists(ctx, key) // 判断 key 是否存在
	if err != nil {
		return false, err
	}
	return exists > 0, nil // exists 是 int64，存在返回 1，不存在返回 0
}

// tokenBlacklistKey 统一生成 token 黑名单 key。
// 将 key 生成逻辑集中起来，避免散落字符串。
func tokenBlacklistKey(token string) string {
	return fmt.Sprintf("jwt:blacklist:%s", token) // 例如 "jwt:blacklist:<token>"
}
