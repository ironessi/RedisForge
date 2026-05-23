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

// DeleteProfileCache 删除用户资料缓存。
// 当用户资料发生变更后调用，保证下次查询能重新从 MySQL 加载最新数据。
func DeleteProfileCache(ctx context.Context, userId uint64) error {
	key := profileCacheKey(userId)

	//Del返回被删除的 key 的数量，成功删除返回 1，key 不存在返回 0。
	_, err := g.Redis().Del(ctx, key)
	return err
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

// UpdateProfile 更新当前用户资料。
// 更新 MySQL 成功后删除 Redis 缓存，让下次查询重新加载最新数据。
func UpdateProfile(ctx context.Context, userId uint64, nickname string) error {
	// 先确认用户是否存在，避免更新不存在的用户。
	count, err := dao.User.Ctx(ctx).Where("id", userId).Count() //查询用户数量
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("用户不存在")
	}

	// 更新 MySQL 中的用户资料。
	_, err = dao.User.Ctx(ctx).Where("id", userId).Data(g.Map{
		"nickname": nickname,
	}).Update()
	if err != nil {
		return err
	}

	// 更新成功后删除 Redis 缓存。
	if err := DeleteProfileCache(ctx, userId); err != nil {
		return err
	}
	return nil
}
