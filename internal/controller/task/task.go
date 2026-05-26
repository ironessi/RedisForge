package task

// ControllerV1 是 task 模块 v1 版本控制器。
type ControllerV1 struct{}

// NewV1 创建 task v1 控制器实例。
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}
