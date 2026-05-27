package notification

// ControllerV1 是 notification 模块 v1 版本控制器。
type ControllerV1 struct{}

// NewV1 创建 notification v1 控制器实例。
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
