package presence

import (
	"context"
	"fmt"
	"redis-demo/internal/dao"
	"redis-demo/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	presenceUserExpire = 60 * time.Second
	presenceTeamExpire = time.Hour
)

// Heartbeat 记录用户在线心跳。
// 只要 presence:user:{userId} 在 Redis 中存在，就认为用户在线。
func Heartbeat(ctx context.Context, userId uint64, teamId uint64) error {
	// 确认用户属于这个团队，避免给无关团队伪造在线状态。
	count, err := dao.TeamMember.Ctx(ctx).Where("user_id", userId).Where("team_id", teamId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("用户不属于这个团队")
	}

	// 用户维度 key：60 秒没有心跳就自动离线。
	if err := g.Redis().SetEX(ctx, presenceUserKey(userId), teamId, int64(presenceUserExpire.Seconds())); err != nil {
		return err
	}

	teamKey := presenceTeamKey(teamId)

	// 团队维度 Set：记录最近在线过的成员，用于快速查询团队在线候选人。
	if _, err := g.Redis().SAdd(ctx, teamKey, userId); err != nil { // 记录最近在线过的成员。
		return err
	}

	// 给团队在线集合设置较长 TTL，避免长期无人访问的团队在线集合一直存在。
	if _, err := g.Redis().Expire(ctx, teamKey, int64(presenceTeamExpire.Seconds())); err != nil {
		return err

	}

	return nil
}

// GetOnlineMembers 查询团队在线成员。
// 它会先查团队成员，再逐个检查 presence:user:{userId} 是否存在。
func GetOnlineMembers(ctx context.Context, userId uint64, teamId uint64) ([]entity.User, error) {
	count, err := dao.TeamMember.Ctx(ctx).
		Where("team_id", teamId).
		Where("user_id", userId).
		Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("你没有权限查看该团队在线成员")
	}

	teamKey := presenceTeamKey(teamId)

	// 查询团队最近在线候选成员。
	values, err := g.Redis().SMembers(ctx, teamKey)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return []entity.User{}, nil
	}

	onlineUserIds := make([]uint64, 0, len(values)) //切片预分配空间，避免多次扩容

	for _, value := range values {
		userId := value.Uint64() //将字符串转换为 uint64。

		// 判断用户在线 key 是否还存在。不存在说明 TTL 已过期，用户已离线。
		exists, err := g.Redis().Exists(ctx, presenceUserKey(userId))
		if err != nil {
			return []entity.User{}, err
		}
		if exists > 0 {
			onlineUserIds = append(onlineUserIds, userId)
			continue
		}

		// 顺手清理团队 Set 中的离线用户，避免脏数据长期堆积。
		if _, err := g.Redis().SRem(ctx, teamKey, userId); err != nil { //SRem() 移除集合中的元素。
			return nil, err
		}
	}

	if len(onlineUserIds) == 0 {
		return []entity.User{}, nil //没有在线成员，返回空列表
	}

	var users []entity.User
	// 查询在线用户的基础信息。
	err = dao.User.Ctx(ctx).WhereIn("id", onlineUserIds).Scan(&users)
	if err != nil {
		return nil, err
	}

	return users, nil

}

// presenceUserKey 生成用户在线状态 key。
// 例如：presence:user:1
func presenceUserKey(userId uint64) string {
	return fmt.Sprintf("presence:user:%d", userId)
}

// presenceTeamKey 生成团队在线集合 key。
// 例如：presence:team:1
func presenceTeamKey(teamId uint64) string {
	return fmt.Sprintf("presence:team:%d", teamId)
}
