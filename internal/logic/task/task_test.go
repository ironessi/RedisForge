package task

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"redis-demo/internal/dao"
	notificationLogic "redis-demo/internal/logic/notification"
	"redis-demo/internal/model/entity"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func TestGetTaskCachesNullForMissingTask(t *testing.T) {
	ctx := gctx.New()
	missingTaskId := uint64(time.Now().UnixNano())
	key := taskDetailCacheKey(missingTaskId)

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean task detail null cache failed: %v", err)
		}
	})

	_, err := GetTask(ctx, 1, missingTaskId)
	if err == nil {
		t.Fatal("missing task should return error")
	}
	if !strings.Contains(err.Error(), "任务不存在") {
		t.Fatalf("unexpected missing task error: %v", err)
	}

	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read null cache failed: %v", err)
	}
	if value.String() != taskDetailCacheNullValue {
		t.Fatalf("unexpected null cache value: %q", value.String())
	}

	ttl, err := g.Redis().TTL(ctx, key)
	if err != nil {
		t.Fatalf("read null cache ttl failed: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("null cache should have ttl, got %d", ttl)
	}

	_, err = GetTask(ctx, 1, missingTaskId)
	if err == nil {
		t.Fatal("missing task cached as null should still return error")
	}
	if !strings.Contains(err.Error(), "任务不存在") {
		t.Fatalf("unexpected cached missing task error: %v", err)
	}
}

func TestGetTaskCachesExistingTaskAndKeepsPermissionCheck(t *testing.T) {
	ctx := gctx.New()
	task, memberId, outsiderId := updateTaskTestFixture(t)
	key := taskDetailCacheKey(task.Id)

	if _, err := g.Redis().Del(ctx, key); err != nil {
		t.Fatalf("clean task detail cache before test failed: %v", err)
	}

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean task detail cache after test failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), -1, task.Id); err != nil {
			t.Errorf("restore heat after member detail read failed: %v", err)
		}
	})

	item, err := GetTask(ctx, memberId, task.Id)
	if err != nil {
		t.Fatalf("member get task failed: %v", err)
	}
	if item.TaskId != task.Id {
		t.Fatalf("unexpected task item: %+v", item)
	}

	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read task detail cache failed: %v", err)
	}
	if value.IsNil() || value.String() == taskDetailCacheNullValue {
		t.Fatalf("task detail cache should contain task json, got %q", value.String())
	}

	var cached entity.Task
	if err := json.Unmarshal([]byte(value.String()), &cached); err != nil {
		t.Fatalf("task detail cache should be json: %v", err)
	}
	if cached.Id != task.Id || cached.TeamId != task.TeamId {
		t.Fatalf("unexpected cached task: %+v", cached)
	}

	_, err = GetTask(ctx, outsiderId, task.Id)
	if err == nil {
		t.Fatal("cached task should still reject outsider")
	}
	if !strings.Contains(err.Error(), "你没有权限查看该任务") {
		t.Fatalf("unexpected outsider error: %v", err)
	}
}

func TestUpdateTaskDeletesTaskDetailCache(t *testing.T) {
	ctx := gctx.New()
	original, operatorId, _ := updateTaskTestFixture(t)
	key := taskDetailCacheKey(original.Id)
	activityKey := fmt.Sprintf("team:activities:%d", original.TeamId)

	if err := g.Redis().SetEX(ctx, key, `{"cached":"old"}`, int64(taskDetailCacheExpire.Seconds())); err != nil {
		t.Fatalf("prepare task detail cache failed: %v", err)
	}

	updatedTitle := original.Title + "-cache-delete"
	updatedDescription := original.Description + "-cache-delete"

	err := UpdateTask(
		ctx,
		operatorId,
		original.Id,
		updatedTitle,
		updatedDescription,
		original.AssigneeId,
		uint(original.Priority),
	)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	t.Cleanup(func() {
		restoreTaskEditableFields(t, original)
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean task detail cache failed: %v", err)
		}
		if _, err := g.Redis().LPop(ctx, activityKey); err != nil {
			t.Errorf("remove test activity failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(original.TeamId), -1, original.Id); err != nil {
			t.Errorf("restore heat failed: %v", err)
		}
	})

	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read task detail cache failed: %v", err)
	}
	if !value.IsNil() {
		t.Fatalf("task detail cache should be deleted after update, got %q", value.String())
	}
}

func TestUpdateStatusDeletesTaskDetailCache(t *testing.T) {
	ctx := gctx.New()
	task, operatorId, _ := statusNotificationTestFixture(t)
	key := taskDetailCacheKey(task.Id)
	activityKey := fmt.Sprintf("team:activities:%d", task.TeamId)
	newStatus := nextStatus(task.Status)

	if err := g.Redis().SetEX(ctx, key, `{"cached":"old"}`, int64(taskDetailCacheExpire.Seconds())); err != nil {
		t.Fatalf("prepare task detail cache failed: %v", err)
	}

	err := UpdateStatus(ctx, operatorId, task.Id, newStatus)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	t.Cleanup(func() {
		restoreTaskForStatusTest(t, task)
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean task detail cache failed: %v", err)
		}
		if _, err := g.Redis().LPop(ctx, activityKey); err != nil {
			t.Errorf("remove test activity failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), -1, task.Id); err != nil {
			t.Errorf("restore heat failed: %v", err)
		}
	})

	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read task detail cache failed: %v", err)
	}
	if !value.IsNil() {
		t.Fatalf("task detail cache should be deleted after status update, got %q", value.String())
	}
}

func TestTaskDetailCacheTTLHasJitterRange(t *testing.T) {
	minTTL := taskDetailCacheExpire
	maxTTL := taskDetailCacheExpire + taskDetailCacheJitterMax

	seenJitter := false
	for i := 0; i < 100; i++ {
		ttl := taskDetailCacheTTL()
		if ttl < minTTL || ttl > maxTTL {
			t.Fatalf("task detail cache ttl out of range: got=%s min=%s max=%s", ttl, minTTL, maxTTL)
		}
		if ttl > minTTL {
			seenJitter = true
		}
	}

	if !seenJitter {
		t.Fatal("task detail cache ttl should include random jitter")
	}
}

func TestGetTaskReleasesDetailLockAfterCacheRebuild(t *testing.T) {
	ctx := gctx.New()
	task, memberId, _ := updateTaskTestFixture(t)
	cacheKey := taskDetailCacheKey(task.Id)
	lockKey := taskDetailLockKey(task.Id)

	if _, err := g.Redis().Del(ctx, cacheKey); err != nil {
		t.Fatalf("clean task detail cache before test failed: %v", err)
	}
	if _, err := g.Redis().Del(ctx, lockKey); err != nil {
		t.Fatalf("clean task detail lock before test failed: %v", err)
	}

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, cacheKey, lockKey); err != nil {
			t.Errorf("clean task detail cache and lock failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), -1, task.Id); err != nil {
			t.Errorf("restore heat after detail read failed: %v", err)
		}
	})

	item, err := GetTask(ctx, memberId, task.Id)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}
	if item.TaskId != task.Id {
		t.Fatalf("unexpected task item: %+v", item)
	}

	value, err := g.Redis().Get(ctx, cacheKey)
	if err != nil {
		t.Fatalf("read rebuilt task detail cache failed: %v", err)
	}
	if value.IsNil() {
		t.Fatal("task detail cache should be rebuilt")
	}

	lockValue, err := g.Redis().Get(ctx, lockKey)
	if err != nil {
		t.Fatalf("read task detail lock after rebuild failed: %v", err)
	}
	if !lockValue.IsNil() {
		t.Fatalf("task detail lock should be released after rebuild, got %q", lockValue.String())
	}
}

func TestUpdateTaskRejectsNonMember(t *testing.T) {
	ctx := gctx.New()
	task, _, outsiderId := updateTaskTestFixture(t)

	err := UpdateTask(
		ctx,
		outsiderId,
		task.Id,
		"不应该被修改的标题",
		"不应该被修改的说明",
		0,
		2,
	)

	if err == nil {
		t.Fatal("UpdateTask should reject non-member")
	}

	if !strings.Contains(err.Error(), "你没有权限修改该任务") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestUpdateTaskRejectsAssigneeOutsideTeam(t *testing.T) {
	ctx := gctx.New()
	task, operatorId, outsiderId := updateTaskTestFixture(t)

	err := UpdateTask(
		ctx,
		operatorId,
		task.Id,
		"不会实际保存的标题",
		"不会实际保存的说明",
		outsiderId,
		2,
	)

	if err == nil {
		t.Fatal("UpdateTask should reject assignee outside team")
	}

	if !strings.Contains(err.Error(), "负责人不是该团队成员") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTaskUnchangedDoesNotAddActivityOrHeat(t *testing.T) {
	ctx := gctx.New()
	current, operatorId, _ := updateTaskTestFixture(t)

	activityKey := fmt.Sprintf("team:activities:%d", current.TeamId)

	beforeScore, err := g.Redis().ZScore(ctx, taskHotKey(current.TeamId), current.Id)
	if err != nil {
		t.Fatalf("read heat before update failed: %v", err)
	}

	beforeCount, err := g.Redis().LLen(ctx, activityKey)
	if err != nil {
		t.Fatalf("read activity count before update failed: %v", err)
	}

	err = UpdateTask(
		ctx,
		operatorId,
		current.Id,
		current.Title,
		current.Description,
		current.AssigneeId,
		uint(current.Priority),
	)
	if err != nil {
		t.Fatalf("UpdateTask unchanged failed: %v", err)
	}

	afterScore, err := g.Redis().ZScore(ctx, taskHotKey(current.TeamId), current.Id)
	if err != nil {
		t.Fatalf("read heat after update failed: %v", err)
	}

	afterCount, err := g.Redis().LLen(ctx, activityKey)
	if err != nil {
		t.Fatalf("read activity count after update failed: %v", err)
	}

	if afterScore != beforeScore {
		t.Fatalf("heat changed: before=%v after=%v", beforeScore, afterScore)
	}
	if afterCount != beforeCount {
		t.Fatalf("activity count changed: before=%v after=%v", beforeCount, afterCount)
	}
}
func TestUpdateTaskUpdatesAndClearsAssignee(t *testing.T) {
	ctx := gctx.New()
	original, operatorId, _ := updateTaskTestFixture(t)

	activityKey := fmt.Sprintf("team:activities:%d", original.TeamId)
	updatedTitle := original.Title + "-test"
	updatedDescription := original.Description + "-test"
	updatedPriority := uint(3)
	if original.Priority == 3 {
		updatedPriority = 2
	}

	err := UpdateTask(
		ctx,
		operatorId,
		original.Id,
		updatedTitle,
		updatedDescription,
		0,
		updatedPriority,
	)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	t.Cleanup(func() {
		var originalAssignee any
		if original.AssigneeId > 0 {
			originalAssignee = original.AssigneeId
		}

		_, err := dao.Task.Ctx(ctx).Where("id", original.Id).Data(g.Map{
			"title":       original.Title,
			"description": original.Description,
			"assignee_id": originalAssignee,
			"priority":    original.Priority,
		}).Update()
		if err != nil {
			t.Errorf("restore task fields failed: %v", err)
		}

		if _, err := g.Redis().LPop(ctx, activityKey); err != nil {
			t.Errorf("remove test activity failed: %v", err)
		}

		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(original.TeamId), -1, original.Id); err != nil {
			t.Errorf("restore heat failed: %v", err)
		}
	})
	var updated entity.Task
	if err := dao.Task.Ctx(ctx).Where("id", original.Id).Scan(&updated); err != nil {
		t.Fatalf("query updated task failed: %v", err)
	}

	if updated.Title != updatedTitle ||
		updated.Description != updatedDescription ||
		updated.AssigneeId != 0 ||
		updated.Priority != int(updatedPriority) {
		t.Fatalf("unexpected updated task: %+v", updated)
	}

}

// updateTaskTestFixture finds one existing task, a member allowed to edit it,
// and a user outside its team. The permission tests remain valid as team data evolves.
func updateTaskTestFixture(t *testing.T) (entity.Task, uint64, uint64) {
	t.Helper()
	ctx := gctx.New()

	var task entity.Task
	if err := dao.Task.Ctx(ctx).OrderAsc("id").Limit(1).Scan(&task); err != nil {
		t.Fatalf("query task fixture failed: %v", err)
	}
	if task.Id == 0 {
		t.Fatal("task fixture does not exist")
	}

	var members []entity.TeamMember
	if err := dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).Scan(&members); err != nil {
		t.Fatalf("query team members failed: %v", err)
	}
	if len(members) == 0 {
		t.Fatalf("team %d has no editable member fixture", task.TeamId)
	}

	memberIds := make(map[uint64]struct{}, len(members))
	for _, member := range members {
		memberIds[member.UserId] = struct{}{}
	}

	var users []entity.User
	if err := dao.User.Ctx(ctx).Scan(&users); err != nil {
		t.Fatalf("query user fixtures failed: %v", err)
	}
	for _, user := range users {
		if _, isMember := memberIds[user.Id]; !isMember {
			return task, members[0].UserId, user.Id
		}
	}

	t.Fatalf("team %d has no outside user fixture", task.TeamId)
	return entity.Task{}, 0, 0
}

func restoreTaskEditableFields(t *testing.T, original entity.Task) {
	t.Helper()
	ctx := gctx.New()

	var originalAssignee any
	if original.AssigneeId > 0 {
		originalAssignee = original.AssigneeId
	}

	_, err := dao.Task.Ctx(ctx).Where("id", original.Id).Data(g.Map{
		"title":       original.Title,
		"description": original.Description,
		"assignee_id": originalAssignee,
		"priority":    original.Priority,
	}).Update()
	if err != nil {
		t.Errorf("restore task editable fields failed: %v", err)
	}
}

func TestUpdateStatusCreatesNotificationForAssignee(t *testing.T) {
	ctx := gctx.New()
	task, operatorId, assigneeId := statusNotificationTestFixture(t)
	newStatus := nextStatus(task.Status)
	activityKey := fmt.Sprintf("team:activities:%d", task.TeamId)

	_, err := dao.Task.Ctx(ctx).Where("id", task.Id).Data(g.Map{
		"assignee_id": assigneeId,
	}).Update()
	if err != nil {
		t.Fatalf("prepare assignee failed: %v", err)
	}

	err = UpdateStatus(ctx, operatorId, task.Id, newStatus)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	var created entity.Notification
	err = dao.Notification.Ctx(ctx).
		Where("receiver_id", assigneeId).
		Where("actor_id", operatorId).
		Where("type", notificationLogic.TypeTaskStatusUpdated).
		Where("related_task_id", task.Id).
		OrderDesc("id").
		Limit(1).
		Scan(&created)
	if err != nil {
		t.Fatalf("query created notification failed: %v", err)
	}
	if created.Id == 0 {
		t.Fatal("status update notification was not created")
	}
	if !strings.Contains(created.Content, newStatus) {
		t.Fatalf("notification content should include new status, got: %s", created.Content)
	}

	inUnreadSet, err := g.Redis().SIsMember(ctx, fmt.Sprintf("notification:unread:%d", assigneeId), created.Id)
	if err != nil {
		t.Fatalf("check unread set failed: %v", err)
	}
	if inUnreadSet != 1 {
		t.Fatalf("notification %d was not added to unread set", created.Id)
	}

	t.Cleanup(func() {
		restoreTaskForStatusTest(t, task)

		if _, err := dao.Notification.Ctx(ctx).Where("id", created.Id).Delete(); err != nil {
			t.Errorf("delete test notification failed: %v", err)
		}
		if _, err := g.Redis().SRem(ctx, fmt.Sprintf("notification:unread:%d", assigneeId), created.Id); err != nil {
			t.Errorf("remove test notification from unread set failed: %v", err)
		}
		if _, err := g.Redis().LPop(ctx, activityKey); err != nil {
			t.Errorf("remove test activity failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), -1, task.Id); err != nil {
			t.Errorf("restore heat failed: %v", err)
		}
	})
}

func TestUpdateStatusUnchangedDoesNotCreateNotification(t *testing.T) {
	ctx := gctx.New()
	task, operatorId, assigneeId := statusNotificationTestFixture(t)
	activityKey := fmt.Sprintf("team:activities:%d", task.TeamId)

	beforeNotifications, err := dao.Notification.Ctx(ctx).
		Where("receiver_id", assigneeId).
		Where("actor_id", operatorId).
		Where("type", notificationLogic.TypeTaskStatusUpdated).
		Where("related_task_id", task.Id).
		Count()
	if err != nil {
		t.Fatalf("count notifications before update failed: %v", err)
	}

	beforeActivities, err := g.Redis().LLen(ctx, activityKey)
	if err != nil {
		t.Fatalf("read activity count before update failed: %v", err)
	}

	beforeHeat, err := g.Redis().ZScore(ctx, taskHotKey(task.TeamId), task.Id)
	if err != nil {
		t.Fatalf("read heat before update failed: %v", err)
	}

	err = UpdateStatus(ctx, operatorId, task.Id, task.Status)
	if err != nil {
		t.Fatalf("UpdateStatus unchanged failed: %v", err)
	}

	afterNotifications, err := dao.Notification.Ctx(ctx).
		Where("receiver_id", assigneeId).
		Where("actor_id", operatorId).
		Where("type", notificationLogic.TypeTaskStatusUpdated).
		Where("related_task_id", task.Id).
		Count()
	if err != nil {
		t.Fatalf("count notifications after update failed: %v", err)
	}

	afterActivities, err := g.Redis().LLen(ctx, activityKey)
	if err != nil {
		t.Fatalf("read activity count after update failed: %v", err)
	}

	afterHeat, err := g.Redis().ZScore(ctx, taskHotKey(task.TeamId), task.Id)
	if err != nil {
		t.Fatalf("read heat after update failed: %v", err)
	}

	if afterNotifications != beforeNotifications {
		t.Fatalf("notification count changed: before=%d after=%d", beforeNotifications, afterNotifications)
	}
	if afterActivities != beforeActivities {
		t.Fatalf("activity count changed: before=%d after=%d", beforeActivities, afterActivities)
	}
	if afterHeat != beforeHeat {
		t.Fatalf("heat changed: before=%v after=%v", beforeHeat, afterHeat)
	}
}

func TestUpdateStatusDoesNotNotifySelf(t *testing.T) {
	ctx := gctx.New()
	task, operatorId, _ := statusNotificationTestFixture(t)
	newStatus := nextStatus(task.Status)
	activityKey := fmt.Sprintf("team:activities:%d", task.TeamId)

	_, err := dao.Task.Ctx(ctx).Where("id", task.Id).Data(g.Map{
		"assignee_id": operatorId,
	}).Update()
	if err != nil {
		t.Fatalf("prepare self assignee failed: %v", err)
	}

	beforeNotifications, err := dao.Notification.Ctx(ctx).
		Where("receiver_id", operatorId).
		Where("actor_id", operatorId).
		Where("type", notificationLogic.TypeTaskStatusUpdated).
		Where("related_task_id", task.Id).
		Count()
	if err != nil {
		t.Fatalf("count notifications before update failed: %v", err)
	}

	err = UpdateStatus(ctx, operatorId, task.Id, newStatus)
	if err != nil {
		t.Fatalf("UpdateStatus self assignee failed: %v", err)
	}

	afterNotifications, err := dao.Notification.Ctx(ctx).
		Where("receiver_id", operatorId).
		Where("actor_id", operatorId).
		Where("type", notificationLogic.TypeTaskStatusUpdated).
		Where("related_task_id", task.Id).
		Count()
	if err != nil {
		t.Fatalf("count notifications after update failed: %v", err)
	}

	if afterNotifications != beforeNotifications {
		t.Fatalf("self notification should not be created: before=%d after=%d", beforeNotifications, afterNotifications)
	}

	t.Cleanup(func() {
		restoreTaskForStatusTest(t, task)

		if _, err := g.Redis().LPop(ctx, activityKey); err != nil {
			t.Errorf("remove test activity failed: %v", err)
		}
		if _, err := g.Redis().ZIncrBy(ctx, taskHotKey(task.TeamId), -1, task.Id); err != nil {
			t.Errorf("restore heat failed: %v", err)
		}
	})
}

func statusNotificationTestFixture(t *testing.T) (entity.Task, uint64, uint64) {
	t.Helper()
	ctx := gctx.New()

	var tasks []entity.Task
	if err := dao.Task.Ctx(ctx).OrderAsc("id").Scan(&tasks); err != nil {
		t.Fatalf("query task fixtures failed: %v", err)
	}

	for _, task := range tasks {
		var members []entity.TeamMember
		if err := dao.TeamMember.Ctx(ctx).Where("team_id", task.TeamId).OrderAsc("user_id").Scan(&members); err != nil {
			t.Fatalf("query team members failed: %v", err)
		}
		if len(members) >= 2 {
			return task, members[0].UserId, members[1].UserId
		}
	}

	t.Fatal("no task fixture with at least two team members")
	return entity.Task{}, 0, 0
}

func restoreTaskForStatusTest(t *testing.T, original entity.Task) {
	t.Helper()
	ctx := gctx.New()

	var originalAssignee any
	if original.AssigneeId > 0 {
		originalAssignee = original.AssigneeId
	}

	_, err := dao.Task.Ctx(ctx).Where("id", original.Id).Data(g.Map{
		"status":      original.Status,
		"assignee_id": originalAssignee,
	}).Update()
	if err != nil {
		t.Errorf("restore task status fields failed: %v", err)
	}
}

func nextStatus(status string) string {
	if status == "doing" {
		return "done"
	}
	return "doing"
}
