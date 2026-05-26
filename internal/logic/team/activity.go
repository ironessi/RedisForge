package team

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "redis-demo/api/team/v1"
	"redis-demo/internal/dao"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	activityMaxCount = 100
	activityExpire   = 7 * 24 * time.Hour
)

// AddActivity 将一条团队动态写入 Redis List。
// 最新动态写入列表头部，并只保留最近 100 条。
func AddActivity(ctx context.Context, teamId uint64, activity v1.ActivityItem) error {
	key := teamActivitiesKey(teamId)

	// 将结构体序列化为 JSON，便于 Redis List 保存完整动态内容。
	bytes, err := json.Marshal(activity) // bytes 是一个字节切片，用于保存动态内容。
	if err != nil {
		return err
	}
	// 写入 Redis List，最新动态写在列表头部。
	if _, err := g.Redis().LPush(ctx, key, string(bytes)); err != nil {
		return err
	}

	// 只保留最近 100 条动态，避免列表无限增长。
	if err := g.Redis().LTrim(ctx, key, 0, activityMaxCount-1); err != nil {
		return err
	}

	// 动态流只用于最近事件展示，7 天后没有新动态则自动过期。
	if _, err := g.Redis().Expire(ctx, key, int64(activityExpire.Seconds())); err != nil {
		return err
	}

	return nil
}

// GetActivities 查询团队最近动态。
// Redis List 按最新在前的顺序返回前 20 条动态。
func GetActivities(ctx context.Context, userId uint64, teamId uint64) ([]v1.ActivityItem, error) {
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", userId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("你没有权限查看该团队的动态")
	}

	key := teamActivitiesKey(teamId)

	values, err := g.Redis().LRange(ctx, key, 0, 19) // 获取列表前 20 条动态。values 是一个字符串切片，每个元素都是一个 JSON 字符串，表示一条动态。
	if err != nil {
		return nil, err
	}

	activities := make([]v1.ActivityItem, 0, len(values)) //
	for _, value := range values {
		var activity v1.ActivityItem // 将 JSON 字符串反序列化为结构体。
		// 把 Redis 读取出的 JSON 字符串
		// 转换成 []byte
		// 再反序列化为 ActivityItem 结构体
		if err := json.Unmarshal([]byte(value.String()), &activity); err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// teamActivitiesKey 统一生成团队动态 Redis key。
// 例如 teamId=1 时，key 为：team:activities:1
func teamActivitiesKey(teamId uint64) string {
	return fmt.Sprintf("team:activities:%d", teamId)
}
