package team

import (
	"context"
	"fmt"
	v1 "redis-demo/api/team/v1"
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
		return gerror.New("用户已是团队成员")
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

// GetMembers 查询团队成员列表。
// 优先使用 Redis Set 中的成员 userId；缓存不存在时从 MySQL 重建。
func GetMembers(ctx context.Context, teamId uint64) ([]v1.MemberItem, error) {
	// 1. 判断团队是否存在。
	count, err := dao.Team.Ctx(ctx).Where("id", teamId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("团队不存在")
	}

	key := teamMembersKey(teamId)
	// 2. 优先从 Redis Set 获取成员 userId 列表。
	values, err := g.Redis().SMembers(ctx, key) //返回字符串切片
	if err != nil {
		return nil, err
	}
	userIds := make([]uint64, 0, len(values)) //创建一个 uint64 类型的切片，长度为 0，容量为 Redis 返回的成员数量
	for _, v := range values {                //v遍历SMembers返回的字符串切片，每个v是一个字符串，表示一个成员的userId
		userIds = append(userIds, v.Uint64()) //将字符串转换为 uint64 并添加到 userIds 切片中
	}

	// 3. 如果 Redis 没有命中，则从 MySQL 查询成员关系并重建缓存。
	if len(userIds) == 0 {
		var members []entity.TeamMember
		err = dao.TeamMember.Ctx(ctx).Where("team_id", teamId).Scan(&members)
		if err != nil {
			return nil, err
		}

		if len(members) == 0 {
			return []v1.MemberItem{}, nil //团队没有成员，返回空列表
		}

		for _, member := range members {
			userIds = append(userIds, member.UserId) //将成员的 userId 添加到 userIds 切片中
		}

		// 将 MySQL 中的成员 ID 重建到 Redis Set。
		args := make([]any, 0, len(userIds))
		for _, id := range userIds {
			args = append(args, id) //将 userIds 中的每个 userId 添加到 args 切片中，准备传递给 SAdd 方法
		}
		if _, err := g.Redis().SAdd(ctx, key, args[0], args[1:]...); err != nil {
			return nil, err
		}

		// 设置缓存 TTL，避免团队成员缓存长期不刷新。
		if _, err := g.Redis().Expire(ctx, key, int64(teamMembersCacheExpire.Seconds())); err != nil {
			return nil, err
		}
	}

	// 4. 查询成员用户信息。
	var users []entity.User
	err = dao.User.Ctx(ctx).WhereIn("id", userIds).Scan(&users) //根据 userIds 切片中的 userId 查询用户信息，并将结果扫描到 users 切片中
	if err != nil {
		return nil, err
	}

	// 5. 查询成员角色。
	var relations []entity.TeamMember
	err = dao.TeamMember.Ctx(ctx).Where("team_id", teamId).WhereIn("user_id", userIds).Scan(&relations) //根据 teamId 和 userIds 查询团队成员关系，并将结果扫描到 relations 切片中.WhereIn是 gorm 的方法，用于生成 SQL 中的 IN 条件，例如 WHERE user_id IN (1, 2, 3)
	if err != nil {
		return nil, err
	}

	roleMap := make(map[uint64]string, len(relations)) //创建一个 map，key 是 userId，value 是角色字符串，容量为 relations 切片的长度
	for _, relation := range relations {
		roleMap[relation.UserId] = relation.Role //将 relations 切片中的每个成员关系的 userId 和角色添加到 roleMap 中，方便后续根据 userId 获取角色
	}

	// 6. 构造返回结果。
	result := make([]v1.MemberItem, 0, len(users))
	for _, user := range users {
		result = append(result, v1.MemberItem{
			UserId:   user.Id,
			Username: user.Username,
			Nickname: user.Nickname,
			Role:     roleMap[user.Id], //根据 userId 从 roleMap 中获取角色字符串，添加到返回结果中
		})
	}

	return result, nil
}
