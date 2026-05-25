package v1

import "github.com/gogf/gf/v2/frame/g"

// ActivityItem 是团队最近动态列表项。
type ActivityItem struct {
	Action       string `json:"action"`  // 动作类型，如 "created_team", "joined_team", "left_team" 等
	ActorId      uint64 `json:"actorId"` // 触发动作的用户ID
	TargetUserId uint64 `json:"targetUserId"`
	Content      string `json:"content"` // 动作描述内容，如 "Alice 创建了团队"、"Bob 加入了团队" 等
	CreatedAt    int64  `json:"createdAt"`
}

// ActivitiesReq 是查询团队最近动态请求。
type ActivitiesReq struct {
	g.Meta `path:"/teams/{teamId}/activities" method:"get" tags:"Team" summary:"查询团队最近动态"`
	TeamId uint64 `json:"teamId" v:"required|min:1#团队ID不能为空|团队ID不合法"`
}

// ActivitiesRes 是查询团队最近动态响应。
type ActivitiesRes struct {
	Activities []ActivityItem `json:"activities"` // 团队最近动态列表
}
