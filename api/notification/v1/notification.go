package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// NotificationItem 是通知列表中的单条通知。
// receiverId 不返回给前端，因为列表只能查询当前登录用户自己的通知。
type NotificationItem struct {
	NotificationId uint64      `json:"notificationId"` //通知ID
	ActorId        uint64      `json:"actorId"`        //触发通知的用户ID
	Type           string      `json:"type"`           //通知类型，例如 "task_assigned"
	Content        string      `json:"content"`
	RelatedTaskId  uint64      `json:"relatedTaskId"` //相关任务ID，如果通知与任务相关的话
	IsRead         uint        `json:"isRead"`        //是否已读，0表示未读，1表示已读
	CreatedAt      *gtime.Time `json:"createdAt"`     //通知创建时间
	ReadAt         *gtime.Time `json:"readAt"`
}

// ListReq 是查询当前用户通知列表的请求。
// 当前用户身份由 JWT 中间件提供，因此请求中不允许传 receiverId。
type ListReq struct {
	g.Meta `path:"/notifications" method:"get" tags:"Notification" summary:"查询当前用户通知列表"`
}

// ListRes 是查询通知列表的响应。
type ListRes struct {
	Notifications []NotificationItem `json:"notifications"`
}

// ReadReq 是标记单条通知已读的请求。
// notificationId 来自路径参数，当前用户身份来自 JWT 上下文。
type ReadReq struct {
	g.Meta         `path:"/notifications/{notificationId}/read" method:"patch" tags:"Notification" summary:"标记通知为已读"`
	NotificationId uint64 `json:"notificationId" v:"required|min:1#请选择要标记为已读的通知|通知ID不合法"`
}

// ReadRes 是标记通知已读的响应。
// 成功时不需要返回额外数据。
type ReadRes struct{}

// UnreadCountReq 是查询当前用户未读通知数量的请求。
// 当前用户身份来自 JWT 上下文，请求中不接受 receiverId，避免查询他人的未读数量。
type UnreadCountReq struct {
	g.Meta `path:"/notifications/unread-count" method:"get" tags:"Notification" summary:"查询当前用户未读通知数量"`
}

// UnreadCountRes 是查询未读通知数量的响应。
type UnreadCountRes struct {
	Count uint64 `json:"count"`
}
