package user

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "redis-demo/api/user/v1"
	"redis-demo/internal/dao"
	"redis-demo/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const profileCacheExpire = 30 * time.Minute

// profileCacheKey 统一生成用户信息缓存 key。
// 例如 userId=1 时，key 为：user:profile:1
func profileCacheKey(userId uint64) string {
	return fmt.Sprintf("user:profile:%d", userId)
}

// GetProfile 获取用户资料。
// 当前先从 MySQL 查询，下一步再加 Redis 缓存。
func GetProfile(ctx context.Context, userId uint64) (*v1.ProfileRes, error) {

	key := profileCacheKey(userId) // 生成 Redis key，例如 "user:profile:1"
	// 先尝试从 Redis 获取用户信息缓存。
	value, err := g.Redis().Get(ctx, key) // 从 Redis 获取缓存,json字符串
	if err != nil {
		return nil, err
	}
	// 2. 如果 Redis 中存在缓存，直接反序列化并返回。
	if !value.IsNil() { // key 存在，说明有缓存
		var profile v1.ProfileRes

		// 反序列化 Redis 中的 JSON 字符串到 profile 结构体。
		if err := json.Unmarshal([]byte(value.String()), &profile); err != nil { // 反序列化失败，可能是数据格式问题，继续往下走数据库查询。

			return nil, err
		}
		return &profile, nil // 成功从 Redis 获取并反序列化，直接返回用户信息
	}

	var user entity.User
	//根据用户ID查询用户信息
	err = dao.User.Ctx(ctx).Where("id", userId).Scan(&user)
	if err != nil {
		return nil, err
	}
	if user.Id == 0 {
		return nil, gerror.New("用户不存在")
	}

	profile := &v1.ProfileRes{
		Nickname: user.Nickname,
		UserId:   user.Id,
		Username: user.Username,
	}
	// 3. 将用户信息序列化为 JSON 字符串，并存入 Redis，设置过期时间。
	bytes, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	//写入 Redis，并设置 TTL，避免缓存长期不更新。
	if err := g.Redis().SetEX(ctx, key, string(bytes), int64(profileCacheExpire.Seconds())); err != nil {
		return nil, err
	}

	return profile, nil
}
