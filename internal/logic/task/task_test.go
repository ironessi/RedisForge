package task

import (
	"fmt"
	"strings"
	"testing"

	"redis-demo/internal/dao"
	"redis-demo/internal/model/entity"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

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
