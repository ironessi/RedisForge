package team

import (
	"context"
	"fmt"
	"redis-demo/internal/dao"
	"redis-demo/internal/model/do"
	"redis-demo/internal/model/entity"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	TeamRoleOwner  = "owner"
	TeamRoleMember = "member"

	teamMembersCacheExpire = 30 * time.Minute
)

// CreateTeam 创建团队，并将创建者加入团队成员表。
// 创建成功后，会把 ownerId 写入 Redis Set，作为团队成员缓存。
func CreateTeam(ctx context.Context, ownerId uint64, name string) (uint64, error) {
	// 1. 插入团队基础信息。
	result, err := dao.Team.Ctx(ctx).Data(do.Team{
		Name:    name,
		OwnerId: ownerId,
	}).Insert()
	if err != nil {
		return 0, err
	}

	// 2. 获取刚刚创建的团队ID。
	teamId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// 3. 将创建者写入团队成员表，角色为 owner。
	_, err = dao.TeamMember.Ctx(ctx).Data(do.TeamMember{
		TeamId: uint64(teamId),
		UserId: ownerId,
		Role:   TeamRoleOwner,
	}).Insert()
	if err != nil {
		return 0, err
	}
	// 4. 将创建者写入 Redis Set，方便后续快速查询团队成员。
	_, err = g.Redis().SAdd(ctx, teamMembersKey(uint64(teamId)), ownerId)
	if err != nil {
		return 0, err
	}

	// 给团队成员缓存设置 TTL，避免缓存长期不刷新。
	_, err = g.Redis().Expire(ctx, teamMembersKey(uint64(teamId)), int64(teamMembersCacheExpire.Seconds()))
	if err != nil {
		return 0, err
	}

	return uint64(teamId), nil
}

// teamMembersKey 统一生成团队成员 Redis key。
// 例如 teamId=1 时，key 为：team:members:1
func teamMembersKey(teamId uint64) string {
	return fmt.Sprintf("team:members:%d", teamId)
}

// AddMember 将指定用户添加到团队。
// 只有团队 owner 可以添加成员，成功后会同步写入 Redis Set。
func AddMember(ctx context.Context, operatorId, teamId, targetUserId uint64) error {
	// 1. 查询团队是否存在。
	var team entity.Team
	err := dao.Team.Ctx(ctx).Where("id", teamId).Scan(&team)
	if err != nil {
		return err
	}
	if team.Id == 0 {
		return gerror.New("团队不存在")
	}

	// 2. 只有团队创建者 owner 才能添加成员。
	if team.OwnerId != operatorId {
		return gerror.New("只有团队创建者才能添加成员")
	}

	// 3. 判断目标用户是否存在。
	count, err := dao.User.Ctx(ctx).Where("id", targetUserId).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New("目标用户不存在")
	}

	// 4. 判断目标用户是否已经在团队中。
	count, err = dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Where("user_id", targetUserId).Count()
	if err != nil {
		return err
	}
	if count != 0 {
		return gerror.New("用户以是团队成员")
	}

	// 5. 写入团队成员关系表。
	_, err = dao.TeamMember.Ctx(ctx).Data(do.TeamMember{
		TeamId: teamId,
		UserId: targetUserId,
		Role:   TeamRoleMember,
	}).Insert()
	if err != nil {
		return err
	}

	// 6. 同步写入 Redis Set。Set 天然去重，重复添加不会产生重复成员。
	key := teamMembersKey(teamId)
	if _, err = g.Redis().SAdd(ctx, key, targetUserId); err != nil {
		return err
	}

	// 7. 设置团队成员缓存 TTL，避免缓存长期不刷新。
	if _, err = g.Redis().Expire(ctx, key, int64(teamMembersCacheExpire.Seconds())); err != nil {
		return err
	}
	return nil

}
