package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	taskV1 "redis-demo/api/task/v1"
	teamV1 "redis-demo/api/team/v1"
	"redis-demo/internal/dao"
	notificationLogic "redis-demo/internal/logic/notification"
	"redis-demo/internal/logic/team"
	"redis-demo/internal/model/do"
	"redis-demo/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const TaskStatusTodo = "todo"

// taskHotKey 生成团队热门任务排行榜 key。
// 例如 teamId=7 时，key 为：team:task:hot:7
func taskHotKey(teamId uint64) string {
	return fmt.Sprintf("team:task:hot:%d", teamId)
}

// CreateTask 创建团队任务。
// 当前用户必须属于该团队；指定负责人时，负责人也必须属于该团队。
func CreateTask(ctx context.Context, creatorId uint64, teamId uint64, title string, description string, assigneeId uint64, priority uint) (uint64, error) {
	// 1. 校验当前用户是否属于团队，团队成员才可以创建任务。
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", creatorId).Count()
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, gerror.New("你不是该团队成员")
	}

	// 2. 未分配负责人时写入 NULL；有负责人时校验其属于该团队。
	var assigneeValue any
	if assigneeId > 0 {
		count, err = dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", assigneeId).Count()
		if err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, gerror.New("负责人不是该团队成员")
		}
		assigneeValue = assigneeId
	}

	// 3. 插入任务记录，初始状态统一为 todo。
	result, err := dao.Task.Ctx(ctx).Data(do.Task{
		TeamId:      teamId,
		Title:       title,
		Description: description,
		CreatorId:   creatorId,
		AssigneeId:  assigneeValue,
		Priority:    priority,
		Status:      TaskStatusTodo,
	}).Insert()
	if err != nil {
		return 0, err
	}

	taskId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 4. 创建任务后写入团队动态流，继续复用 Redis List。
	if err := team.AddActivity(ctx, teamId, teamV1.ActivityItem{
		Action:    "task_created",
		ActorId:   creatorId,
		Content:   fmt.Sprintf("用户%d创建了任务：%s", creatorId, title),
		CreatedAt: time.Now().Unix(),
	}); err != nil {
		return 0, err
	}

	return uint64(taskId), nil
}

func GetTasks(ctx context.Context, userId uint64, teamId uint64) ([]taskV1.TaskItem, error) {
	// 1. 校验当前用户是否属于团队
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", userId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("你没有权限查看该团队的任务")
	}
	// 2. 查询这个团队下的任务记录
	var tasks []entity.Task
	err = dao.Task.Ctx(ctx).Where("team_id", teamId).OrderDesc("id").Scan(&tasks)
	if err != nil {
		return nil, err
	}
	// 3. 把 entity.Task 转换成 taskV1.TaskItem
	items := make([]taskV1.TaskItem, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, taskV1.TaskItem{
			TaskId:      task.Id,
			CreatorId:   task.CreatorId,
			Title:       task.Title,
			Description: task.Description,
			AssigneeId:  task.AssigneeId,
			Priority:    task.Priority,
			Status:      task.Status,
		})
	}
	// 4. 返回结果
	return items, nil

}

func GetTask(ctx context.Context, userId uint64, taskId uint64) (*taskV1.TaskItem, error) {
	// 1. 根据 taskId 查询任务
	var task entity.Task
	err := dao.Task.Ctx(ctx).Where("id", taskId).Scan(&task)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gerror.New("任务不存在")
	}
	if err != nil {
		return nil, err
	}
	// 2. 判断任务是否存在
	if task.Id == 0 {
		return nil, gerror.New("任务不存在")
	}
	// 3. 校验当前用户是否属于任务所在团队
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).Where("user_id", userId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("你没有权限查看该任务")
	}

	if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), 1, task.Id); err != nil { // 访问 Redis 记录任务热度，方便后续实现热门任务排行榜。
		return nil, err
	}

	// 4. 转换为 taskV1.TaskItem 并返回
	return &taskV1.TaskItem{
		TaskId:      task.Id,
		CreatorId:   task.CreatorId,
		Title:       task.Title,
		Description: task.Description,
		AssigneeId:  task.AssigneeId,
		Priority:    task.Priority,
		Status:      task.Status,
	}, nil
}

// UpdateTask 更新任务的可编辑信息；任务状态由 UpdateStatus 单独修改。
func UpdateTask(ctx context.Context, operatorId uint64, taskId uint64, title string, description string, assigneeId uint64, priority uint) error {
	var task entity.Task
	err := dao.Task.Ctx(ctx).Where("id", taskId).Scan(&task)
	if errors.Is(err, sql.ErrNoRows) {
		return gerror.New("任务不存在")
	}
	if err != nil {
		return err
	}
	if task.Id == 0 {
		return gerror.New("任务不存在")
	}

	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).Where("user_id", operatorId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("你没有权限修改该任务")
	}

	var assigneeValue any
	if assigneeId > 0 {
		count, err = dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).Where("user_id", assigneeId).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			return gerror.New("负责人不是该团队成员")
		}
		assigneeValue = assigneeId
	}

	if task.Title == title &&
		task.Description == description &&
		task.AssigneeId == assigneeId &&
		task.Priority == int(priority) {
		return nil
	}

	_, err = dao.Task.Ctx(ctx).Where("id", taskId).Data(g.Map{
		"title":       title,
		"description": description,
		"assignee_id": assigneeValue,
		"priority":    priority,
	}).Update()
	if err != nil {
		return err
	}

	if err := team.AddActivity(ctx, task.TeamId, teamV1.ActivityItem{
		Action:    "task_updated",
		ActorId:   operatorId,
		Content:   fmt.Sprintf("用户%d更新了任务：%s", operatorId, title),
		CreatedAt: time.Now().Unix(),
	}); err != nil {
		return err
	}

	if _, err = g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), 1, task.Id); err != nil {
		return err
	}

	// 2. 判断负责人是否发生变化，且新负责人非零
	if assigneeId > 0 && assigneeId != task.AssigneeId {
		if err := notificationLogic.CreateNotification(ctx, assigneeId, operatorId, notificationLogic.TypeTaskAssigned, fmt.Sprintf("用户%d将任务%s交给了您", operatorId, title), task.Id); err != nil {
			return err
		}
	}

	// 4. 返回 nil
	return nil
}

func UpdateStatus(ctx context.Context, operatorId uint64, taskId uint64, status string) error {
	// 1. 查询任务是否存在
	var task entity.Task
	err := dao.Task.Ctx(ctx).Where("id", taskId).Scan(&task)
	if errors.Is(err, sql.ErrNoRows) {
		return gerror.New("任务不存在")
	}
	if err != nil {
		return err
	}
	if task.Id == 0 {
		return gerror.New("任务不存在")
	}
	// 2. 校验操作者是否属于任务所在团队
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).Where("user_id", operatorId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("你没有权限修改该任务")
	}

	// 3. 如果状态没变化，直接返回
	if task.Status == status {
		return nil
	}

	// 4. 更新任务状态
	_, err = dao.Task.Ctx(ctx).Where("id", taskId).Data(do.Task{
		Status: status,
	}).Update()
	if err != nil {
		return err
	}
	// 5. 写入团队动态
	if err := team.AddActivity(ctx, task.TeamId, teamV1.ActivityItem{
		Action:    "task_status_updated",
		ActorId:   operatorId,
		Content:   fmt.Sprintf("用户%d将任务%s的状态更新为%s", operatorId, task.Title, status),
		CreatedAt: time.Now().Unix(),
	}); err != nil {
		return err
	}

	// 6. 更新任务热度
	if _, err = g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), 1, task.Id); err != nil {
		return err
	}
	// 7. 如果任务有负责人，且负责人不是操作者本人，则创建状态变化通知
	if task.AssigneeId > 0 && task.AssigneeId != operatorId {
		if err := notificationLogic.CreateNotification(ctx, task.AssigneeId, operatorId, notificationLogic.TypeTaskStatusUpdated, fmt.Sprintf("用户%d将任务%s的状态更新为%s", operatorId, task.Title, status), task.Id); err != nil {
			return err
		}
	}
	// 8. 返回 nil
	return nil
}

func GetHotTasks(ctx context.Context, userId uint64, teamId uint64) ([]taskV1.HotTaskItem, error) {
	// 1. 校验当前用户属于团队
	count, err := dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", userId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("你没有权限查看该团队的热门任务")
	}

	// 2. 从 Redis 读取热度最高的前 10 个 taskId
	values, err := g.Redis().ZRevRange(ctx, taskHotKey(teamId), 0, 9)
	if err != nil {
		return nil, err
	}

	taskIds := values.Uints() // 把 Redis 读取出的字符串切片转换成 uint64 切片，得到热度最高的前 10 个 taskId。
	if len(taskIds) == 0 {
		return []taskV1.HotTaskItem{}, nil
	}
	// 3. 根据 taskId 查询 MySQL 中的任务
	items := make([]taskV1.HotTaskItem, 0, len(taskIds))
	for _, taskId := range taskIds {
		var task entity.Task
		err := dao.Task.Ctx(ctx).Where("id", taskId).Scan(&task)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		score, err := g.Redis().ZScore(ctx, taskHotKey(teamId), taskId) // 从 Redis 读取该 taskId 的热度分数，方便后续返回给前端展示。
		if err != nil {
			return nil, err
		}

		// 将 MySQL 中的任务信息和 Redis 中的热度分数组装成排行榜项。
		items = append(items, taskV1.HotTaskItem{
			TaskId:    task.Id,
			Title:     task.Title,
			Status:    task.Status,
			Priority:  task.Priority,
			ViewCount: uint64(score),
		})

	}
	// 4. 组装带 viewCount 的排行榜结果

	return items, nil
}
