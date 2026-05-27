# TeamPulse 项目开发路线

更新日期：2026-05-26
当前仓库名：`redis-demo`
产品方向：公司内部团队协作与任务流转后端

## 1. 项目目标

TeamPulse 不是秒杀演示项目，而是一个更贴近公司内部系统的 Redis 实战项目。员工登录系统后，可以加入团队、看到在线同事、创建和处理任务、接收通知、浏览团队动态，并查看热门任务。

项目最终覆盖的业务场景：

- 员工注册、登录、鉴权与退出登录
- 团队创建、成员管理与组织数据缓存
- 团队成员在线状态
- 任务/工单创建、查看与流转
- 通知中心与消息已读/未读
- 团队操作动态流
- 热门任务排行
- 关键接口限流
- 后续增强：分布式锁、延迟任务

Redis 在本项目中的角色不是主数据库，而是支撑公司系统中最常见的高频能力：

```text
登录态失效、短期验证码、热点缓存、在线状态、未读数、
实时动态、排行榜、限流、锁与延迟任务
```

## 2. 技术与分层原则

### 技术栈

| 技术 | 用途 |
| --- | --- |
| GoFrame v2 | HTTP 服务、路由绑定、校验、DAO |
| MySQL | 用户、团队、成员、任务、通知等可靠主数据 |
| Redis | 缓存、TTL 状态、动态流、未读集合、排行、限流 |
| JWT | 登录后的 API 身份凭证 |

### 代码分层

| 层级 | 职责 |
| --- | --- |
| `api/` | 定义路由、请求结构、响应结构和参数校验 |
| `internal/controller/` | 获取 JWT 用户信息，调用 Logic，返回响应 |
| `internal/logic/` | 业务规则、权限校验、MySQL 与 Redis 操作 |
| `internal/dao/` | 数据库访问对象 |
| `internal/model/` | MySQL 对应的数据结构 |
| `internal/middleware/` | JWT 鉴权与请求上下文注入 |

### 数据原则

- MySQL 是业务真实数据源，任务和通知不能只存在 Redis。
- Redis 保存可重建、需要 TTL 或需要高性能访问的数据。
- 团队权限必须在 Logic 层校验，不能只依赖 Controller。
- 每个新功能都要验证成功路径和越权路径。

## 3. 当前开发进度总览

| 阶段 | 模块 | 状态 |
| --- | --- | --- |
| 第一阶段 | 用户与团队 | 基础功能与成员列表读取权限已完成验收 |
| 第二阶段 | 在线状态 | 基础功能与在线成员读取权限已完成验收 |
| 第三阶段 | 任务/工单流转 | 主线功能已完成手工验收，任务编辑关键规则已有可重复运行的自动测试 |
| 第四阶段 | 通知中心与未读数 | 指派通知、列表、已读和未读数链路已验收，待状态通知 |
| 第五阶段 | 团队动态流 | 基础功能已完成，正在完成权限验收 |
| 第六阶段 | 热门任务排行 | 基础功能已完成，控制台页面已与任务工作台联动 |
| 第七阶段 | 接口限流 | 未开始 |
| 增强阶段 | 锁、延迟任务、更多组织缓存 | 未开始 |

## 4. Redis Key 总设计

| 业务 | Key 格式 | 类型 | TTL / 说明 | 状态 |
| --- | --- | --- | --- | --- |
| 登录验证码 | `auth:captcha:{username}` | String | 5 分钟 | 已完成 |
| JWT 退出黑名单 | `jwt:blacklist:{token}` | String | 与 token 有效期配合 | 已完成 |
| 用户资料缓存 | `user:profile:{userId}` | String(JSON) | 30 分钟 | 已完成 |
| 团队成员缓存 | `team:members:{teamId}` | Set | 30 分钟 | 已完成 |
| 用户在线状态 | `presence:user:{userId}` | String | 60 秒 | 已完成 |
| 团队在线候选成员 | `presence:team:{teamId}` | Set | 1 小时 | 已完成 |
| 团队动态流 | `team:activities:{teamId}` | List | 7 天，只保留最近 100 条 | 已完成 |
| 通知未读集合 | `notification:unread:{userId}` | Set | 可由 MySQL 重建 | 已实现并验收 |
| 任务热门排行 | `team:task:hot:{teamId}` | Sorted Set | 按详情浏览加分 | 已完成基础功能 |
| 创建任务限流 | `rate:task:create:{userId}:{minute}` | String Counter | 1 分钟 | 待开发 |
| 登录限流 | `rate:login:{ip}:{minute}` | String Counter | 1 分钟 | 待开发 |
| 延迟提醒 | 待定 | Sorted Set / Stream | 增强阶段 | 待开发 |
| 分布式锁 | `lock:{business}:{id}` | String NX | 短 TTL | 待开发 |

说明：团队动态流在当前代码中使用的是 `team:activities:{teamId}`，后续延续这个已实现命名。

## 5. 第一阶段：用户与团队

### 目标

完成员工身份与团队关系的基础能力，为所有后续模块提供登录用户和团队权限边界。

### 已实现接口

| Method | Path | 用途 | 状态 |
| --- | --- | --- | --- |
| `POST` | `/auth/register` | 注册员工账号 | 已完成 |
| `POST` | `/auth/captcha` | 获取登录验证码 | 已完成 |
| `POST` | `/auth/login` | 登录并签发 JWT | 已完成 |
| `POST` | `/auth/logout` | 退出并拉黑当前 JWT | 已完成 |
| `GET` | `/user/profile` | 获取当前员工资料 | 已完成 |
| `PUT` | `/user/profile` | 更新当前员工资料 | 已完成 |
| `POST` | `/teams` | 创建团队 | 已完成 |
| `POST` | `/teams/{teamId}/members` | owner 添加成员 | 已完成 |
| `GET` | `/teams/{teamId}/members` | 查询团队成员 | 已完成基础功能 |

### Redis 学习点

#### 用户资料缓存

```text
GET /user/profile
-> 先读 user:profile:{userId}
-> 缓存未命中时查 MySQL
-> 将结果写回 Redis

PUT /user/profile
-> 更新 MySQL
-> 删除 user:profile:{userId}
```

这是典型 Cache Aside 模式。

#### 团队成员缓存

```text
team:members:{teamId} -> Set(userId)
```

团队创建和添加成员时同步维护 Set，查询成员时可以优先使用缓存中的成员 ID。

### 待收尾事项

`GET /teams/{teamId}/members` 已将当前用户传入 Logic，并在读取成员缓存前校验其属于目标团队；已验证成员可查询、非成员被拒绝。

## 6. 第二阶段：团队成员在线状态

### 目标

模拟公司 IM 或协作平台中的在线成员展示：用户停留在某个团队工作区时定时上报心跳，其他成员可以看到当前在线同事。

### 已实现接口

| Method | Path | 用途 | 状态 |
| --- | --- | --- | --- |
| `POST` | `/presence/heartbeat` | 上报当前用户在线心跳 | 已完成 |
| `GET` | `/teams/{teamId}/online-members` | 查询团队在线成员 | 已完成基础功能 |

### Redis 设计

前端每隔约 30 秒发送一次心跳：

```text
presence:user:{userId} -> teamId
TTL: 60 秒
```

只要该 key 存在，就认为用户在线。

同时用 Set 维护某个团队最近出现过的在线候选成员：

```text
presence:team:{teamId} -> Set(userId)
TTL: 1 小时
```

查询在线成员时：

```text
读取团队 Set
-> 逐个检查 presence:user:{userId} 是否存在
-> 返回仍在线的成员
-> 顺手从 Set 中清理已离线成员
```

### 待收尾事项

`GET /teams/{teamId}/online-members` 已将当前用户传入 Logic，并在读取在线集合前校验其属于目标团队；已验证成员可查询、非成员被拒绝。

## 7. 第三阶段：任务/工单流转

### 目标

完成协作系统的核心实体：团队成员可以创建任务、查看任务、推动任务状态变化，并为动态流、通知和排行提供业务事件。

### 当前已完成

| Method | Path | 用途 | 状态 |
| --- | --- | --- | --- |
| `POST` | `/teams/{teamId}/tasks` | 创建任务 | 已完成并手工验证 |
| `GET` | `/teams/{teamId}/tasks` | 查询团队任务列表 | 已完成并手工验证 |
| `GET` | `/tasks/{taskId}` | 查询任务详情并累计浏览热度 | 已完成并手工验证 |
| `PUT` | `/tasks/{taskId}` | 更新任务信息 | 已完成并手工验证 |
| `PATCH` | `/tasks/{taskId}/status` | 更新任务状态 | 状态流转与热度计分已验证 |
| `GET` | `/teams/{teamId}/tasks/hot` | 查询热门任务排行 | 已完成并手工验证 |

### 当前业务流程

```text
创建任务
-> MySQL 生成 task，初始 status=todo
-> Redis 团队动态新增 task_created

更新任务状态为 doing
-> MySQL 中 status 变化
-> Redis 团队动态新增 task_status_updated
-> Redis Sorted Set 为该任务增加一次热度

查看任务详情
-> 校验访问者属于任务所在团队
-> Redis Sorted Set 为该任务增加浏览热度

查询热门任务
-> Redis 按热度倒序返回任务 ID
-> MySQL 补充任务摘要信息
-> 前端控制台展示热度排行
```

此前已验证非团队成员不能：

- 查看团队任务列表
- 修改任务状态

### 主线接口验收结果

| Method | Path | 用途 | 验收要点 |
| --- | --- | --- | --- |
| `PUT` | `/tasks/{taskId}` | 更新任务信息 | 成员编辑、解除分配、负责人归属和非成员越权均已验证 |

### `GET /tasks/{taskId}` 的实现结果

1. API 已定义 `DetailReq` 与 `DetailRes`。
2. Logic 根据 `taskId` 查任务，并将 `sql.ErrNoRows` 转为“任务不存在”。
3. Logic 使用数据库中的 `task.TeamId` 校验当前用户属于团队。
4. 权限通过后，通过 `ZINCRBY team:task:hot:{teamId} 1 taskId` 累加热度。
5. 已验证成员可查看、非成员被拒绝、查看详情后排行分数增加。

### `PUT /tasks/{taskId}` 的实现规则

建议允许编辑：

```json
{
  "title": "新的任务标题",
  "description": "新的说明",
  "assigneeId": 11,
  "priority": 3
}
```

规则：

- 任务必须存在。
- 操作者必须属于任务所在团队。
- `assigneeId = 0` 表示暂不分配负责人。
- 新负责人非零时必须属于同一团队。
- 状态更新继续使用已经完成的 `PATCH /tasks/{taskId}/status`。
- 编辑成功后向团队动态流追加一条事件。

实现结果：

- API 已定义 `UpdateReq` 与 `UpdateRes`，只允许提交标题、描述、负责人和优先级。
- Logic 使用任务所属团队校验操作者与负责人，不允许跨团队编辑或指派。
- `assigneeId = 0` 时将数据库中的负责人字段写为 `NULL`。
- 提交的字段没有变化时保持幂等，不重复记录动态或增加热度。
- 编辑成功后写入 `task_updated` 团队动态并增加一次排行热度；成员编辑流程已手工验证。
- 控制台已加入任务创建、列表、详情编辑与状态切换入口，并在操作后刷新动态流和热门排行。

手工验收记录（2026-05-27）：

- 团队成员通过控制台成功编辑任务，并通过 `assigneeId = 0` 解除负责人。
- 对无效负责人提交编辑请求被拒绝，负责人必须属于任务所在团队的规则生效。
- 任务状态更新成功写入动态并增加热度；连续查看任务详情也成功增加热度。
- 非团队成员编辑任务被拒绝，编辑接口的团队成员权限校验已验证。

自动测试记录（2026-05-27）：

- `TestUpdateTaskRejectsNonMember` 已通过，覆盖非团队成员编辑失败。
- `TestUpdateTaskRejectsAssigneeOutsideTeam` 已通过，覆盖跨团队负责人失败。
- `TestUpdateTaskUnchangedDoesNotAddActivityOrHeat` 已通过，覆盖无变化提交不追加动态或热度。
- `TestUpdateTaskUpdatesAndClearsAssignee` 已通过，覆盖成功写入编辑字段与将负责人写为空值；页面验收已覆盖从已分配到解除负责人的操作流程。
- 写入测试已在清理阶段恢复任务字段、移除测试动态并回退热度；使用 `-count=2` 连续运行验证通过。

## 8. 第四阶段：通知中心与已读未读

### 目标

将任务流转产生的事件送达到相关员工，例如任务被分配、状态变化、有人提及自己等。

### 计划接口

| Method | Path | 用途 |
| --- | --- | --- |
| `GET` | `/notifications` | 查询当前用户通知列表 | 已实现并验收 |
| `PATCH` | `/notifications/{notificationId}/read` | 标记当前用户的一条通知已读 | 已实现并验收 |
| `GET` | `/notifications/unread-count` | 查询当前用户未读数 | 已实现并验收 |

第一期不提供 `POST /notifications`。通知不是用户随意发布的消息，而是由任务业务动作在 Logic 内部产生，避免伪造通知或绕过权限规则。

### 数据设计

MySQL 保存通知主数据，是通知是否存在、属于谁以及是否已读的真实数据源：

```text
notification
- id               bigint unsigned primary key
- receiver_id      bigint unsigned not null       接收通知的用户
- actor_id         bigint unsigned not null       触发动作的用户
- type             varchar(32) not null           通知类型
- content          varchar(255) not null          展示文案
- related_task_id  bigint unsigned null           关联任务
- is_read          tinyint not null default 0     0 未读，1 已读
- created_at       datetime not null
- read_at          datetime null
```

建议索引：

```text
index(receiver_id, created_at)
index(receiver_id, is_read)
```

第一期建表 SQL：

```sql
CREATE TABLE `notification` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '通知ID',
  `receiver_id` bigint unsigned NOT NULL COMMENT '接收人ID',
  `actor_id` bigint unsigned NOT NULL COMMENT '触发人ID',
  `type` varchar(32) NOT NULL COMMENT '通知类型',
  `content` varchar(255) NOT NULL COMMENT '通知内容',
  `related_task_id` bigint unsigned DEFAULT NULL COMMENT '关联任务ID',
  `is_read` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否已读：0未读/1已读',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `read_at` datetime DEFAULT NULL COMMENT '已读时间',
  PRIMARY KEY (`id`),
  KEY `idx_notification_receiver_created` (`receiver_id`, `created_at`),
  KEY `idx_notification_receiver_read` (`receiver_id`, `is_read`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户通知';
```

第一期通知类型：

| Type | 触发场景 | 接收人 |
| --- | --- | --- |
| `task_assigned` | 任务负责人变更为某成员 | 新负责人 |
| `task_status_updated` | 任务状态发生变化 | 当前负责人 |

Redis 使用 Set 加速未读数量。Set 成员为 `notificationId`，可由 MySQL 中 `is_read = 0` 的数据重建：

```redis
SADD  notification:unread:1 1001
SREM  notification:unread:1 1001
SCARD notification:unread:1
```

### 业务规则

- 通知只能由业务 Logic 创建，不开放前端创建通知接口。
- 用户只能查询自己的通知、读取自己的未读数、将自己的通知标记为已读。
- 通知列表第一期按 `created_at` 倒序返回最近 20 条，避免无边界读取。
- `receiver_id = actor_id` 时不创建通知，避免自己通知自己。
- 任务重新分配时只通知新的非零负责人；解除负责人不创建指派通知。
- 状态更新只通知当前非零负责人，且状态实际变化时才触发。
- 重复标记同一条已读通知应幂等成功，不重复改变 MySQL 或 Redis 计数。
- MySQL 是真实数据源；如果 Redis 未读集合写入失败，已有通知仍以 MySQL 记录为准，并允许查询未读数时重建缓存。

### 业务流程

任务指派通知：

```text
UpdateTask 中 assigneeId 从旧值变为新值
-> 新 assigneeId 非零且不等于 operatorId
-> 插入 MySQL notification(type=task_assigned, receiver_id=新负责人)
-> SADD notification:unread:{新负责人} notificationId
```

任务状态通知：

```text
UpdateStatus 中 status 实际发生变化
-> 当前任务 assigneeId 非零且不等于 operatorId
-> 插入 MySQL notification(type=task_status_updated, receiver_id=负责人)
-> SADD notification:unread:{负责人} notificationId
```

查询通知：

```text
从 JWT 获取当前 userId
-> MySQL 按 receiver_id = userId 查询
-> created_at 倒序返回最近 20 条通知
```

标记已读：

```text
从 JWT 获取当前 userId
-> 只查询 receiver_id = userId 且 id = notificationId 的通知
-> 已经 is_read = 1 时不再更新 MySQL，但仍 SREM 修复可能残留的未读集合成员
-> 更新 MySQL is_read = 1、read_at = 当前时间
-> SREM notification:unread:{userId} notificationId
```

查询未读数：

```text
从 JWT 获取当前 userId
-> SCARD notification:unread:{userId}
-> 数量大于 0 时直接返回
-> 数量为 0 时从 MySQL 查询未读通知 ID，存在未读则回填 Set 后返回数量
-> MySQL 同样无未读时返回 0
```

第一期采用上述简单重建方式，因此“确实没有未读”时会查询一次 MySQL；后续如需减少空结果查询，可增加缓存初始化标记 key。

### 与任务模块的联动

第一期在以下既有业务方法成功写入任务数据后创建通知：

| Logic 方法 | 触发通知 | 说明 |
| --- | --- | --- |
| `UpdateTask` | `task_assigned` | 已接入并验收：负责人发生变化且新负责人不是操作者本人 |
| `UpdateStatus` | `task_status_updated` | 待接入：状态发生变化且负责人不是操作者本人 |

后续增加评论或提及时，再扩展新的通知类型。

### 联调验收记录（2026-05-27）

- A 将任务负责人改为 B 后，B 的 `GET /notifications` 返回未读 `task_assigned` 通知。
- 新通知 ID 已写入 Redis DB 1 的 `notification:unread:{B}` Set。
- A 尝试标记 B 的通知已读被拒绝，返回“你没有权限操作该通知”。
- B 标记自己的通知已读后，列表返回 `isRead = 1` 且 `readAt` 非空，Redis Set 已移除该通知 ID。
- 重复标记已读保持幂等，Redis 未读集合不会恢复脏数据。
- `GET /notifications/unread-count` 已验证无未读时返回 `0`，Redis 已有未读集合时直接返回集合数量。
- 已验证 Redis 集合为空但 MySQL 存在未读通知时，接口返回真实数量并将通知 ID 回填至 `notification:unread:{userId}`。
- 已验证未携带 JWT 时未读数接口返回 `401`，标记全部通知已读后接口和 Redis `SCARD` 均返回 `0`。

### 第一期实现顺序

1. 在 MySQL 创建 `notification` 表，并通过 `make dao` 生成 `dao`、`do`、`entity` 模型。已完成。
2. 定义通知列表与标记已读的 API 请求和响应结构。已完成。
3. 实现内部创建通知、查询列表与标记已读 Logic。已完成。
4. 实现列表、已读 Controller 并注册受 JWT 保护的 notification 模块路由。已完成。
5. 在 `UpdateTask` 中触发 `task_assigned`，并联调创建、读取、已读与 Redis 集合。已完成。
6. 实现并联调未读数接口，验证 Redis 命中、MySQL 重建、JWT 保护与已读清零链路。已完成。
7. 在 `UpdateStatus` 中接入状态通知并联调。待完成。

## 9. 第五阶段：团队操作动态流

### 目标

提供公司协作系统首页常见的“团队最近发生了什么”信息流。

示例：

```text
张三创建了任务 A
李四将任务 B 更新为完成
王五后续评论了工单 C
```

### 当前状态

| 能力 | 状态 |
| --- | --- |
| Redis List 写入与读取 | 已完成 |
| 创建团队写动态 | 已完成 |
| 添加成员写动态 | 已完成 |
| 创建任务写动态 | 已完成 |
| 更新任务状态写动态 | 已完成 |
| 查看动态的成员权限 | 代码已完成，需补一次手工越权验收 |

### 当前接口

```http
GET /teams/{teamId}/activities
```

### Redis 设计

```redis
LPUSH team:activities:{teamId} json
LTRIM team:activities:{teamId} 0 99
LRANGE team:activities:{teamId} 0 19
EXPIRE team:activities:{teamId} 604800
```

### 后续扩展事件

任务完整更新和通知完成后，继续增加：

- `task_updated`
- `task_assigned`
- `task_commented`

## 10. 第六阶段：热门任务排行

### 目标

展示某个团队最近最受关注的任务，这是 Redis Sorted Set 在公司业务中的典型用法。

### 已实现接口

```http
GET /teams/{teamId}/tasks/hot
```

### Redis 设计

```text
key:    team:task:hot:{teamId}
member: taskId
score:  热度分数
```

最初的加分规则建议保持简单：

| 行为 | 分数 |
| --- | --- |
| 有权限的成员查看任务详情 | `+1` |
| 更新任务基本信息 | `+1` |
| 更新任务状态 | `+1` |

实现示意：

```redis
ZINCRBY team:task:hot:7 1 1001
ZREVRANGE team:task:hot:7 0 9 WITHSCORES
```

### 开发依赖

热门排行应在 `GET /tasks/{taskId}` 完成后开发，因为详情浏览是最自然的热度来源。

### 验收场景

```text
查看任务 A 三次
查看任务 B 一次
GET /teams/{teamId}/tasks/hot
-> A 排在 B 之前，并返回对应分数
```

当前已经验证单任务热度自增流程：

```text
读取热门排行，任务 1 热度为 2
-> 请求 GET /tasks/1
-> 再次读取热门排行，任务 1 热度为 3
```

控制台首页已加入热门任务面板，可以输入团队 ID 刷新排行，并通过“查看详情”按钮触发一次热度增长。

## 11. 第七阶段：接口限流

### 目标

保护容易被频繁调用或滥用的接口，练习 Redis 计数器与过期时间的组合。

### 首批限流规则

| 场景 | 规则 | Key |
| --- | --- | --- |
| 创建任务 | 每个用户每分钟最多 10 次 | `rate:task:create:{userId}:{minute}` |
| 登录尝试 | 每个 IP 每分钟最多 5 次 | `rate:login:{ip}:{minute}` |

### Redis 实现思路

```text
INCR key
首次计数时 EXPIRE key 60
若计数超过阈值，则拒绝请求
```

### 放置位置

- 登录限流适合在认证业务入口处理，因为此时还没有 JWT 用户。
- 创建任务限流可以在认证后的 Controller 或 Logic 入口处理。
- 限流逻辑稳定后，可提炼为可配置的中间件或公共 Logic 工具。

### 验收场景

```text
一分钟内连续创建任务 10 次 -> 成功
第 11 次 -> 返回请求过于频繁
等待窗口结束 -> 再次允许创建
```

## 12. 增强阶段：更完整的公司协作能力

完成前七阶段后，可以继续扩展：

| 能力 | 使用 Redis 的方式 | 业务例子 |
| --- | --- | --- |
| 组织/团队详情缓存 | String(JSON) Cache Aside | 高频查看组织页 |
| 分布式锁 | `SET key value NX EX` | 防止重复处理任务或重复发送通知 |
| 延迟任务 | Sorted Set / 队列 | 任务到期提醒、通知重试 |
| 工单评论 | 动态流 + 通知 | 评论任务并提醒负责人 |
| 过期任务提醒 | 延迟任务 + 通知 | 截止时间前提醒成员 |

这些功能不应抢在核心流程之前实现；先让任务、通知、动态、排行和限流形成完整闭环。

## 13. 推荐实施顺序

| 顺序 | 功能 | 原因 |
| --- | --- | --- |
| 1 | 实现通知中心与未读数 | 让任务事件真正触达员工 |
| 2 | 完善通知与动态自动测试 | 固化消息触达和信息流规则 |
| 3 | 完善任务与动态自动测试 | 继续固化已实现的业务规则 |
| 4 | 完善动态事件与手工验收 | 让工作台信息流完整 |
| 5 | 接口限流 | 为关键写接口和登录增加保护 |
| 6 | 自动测试、README 与演示脚本 | 形成可展示项目 |
| 7 | 锁与延迟任务 | 作为进阶能力 |

## 14. 里程碑与验收标准

### M1：团队协作基础

状态：已完成

- 员工可注册、登录、退出。
- 可创建团队并添加成员。
- 用户资料与团队成员使用 Redis 缓存。
- 在线心跳和动态查询基本可用。
- 成员、在线成员读取权限已补齐并完成成员/非成员手工验收。

### M2：任务流转与热度

状态：进行中

- 创建、列表、详情、状态流转、热门排行和基本编辑均已实现并完成主线手工验收，任务编辑关键规则已有可重复运行的自动测试覆盖。
- 所有任务读取和修改都有团队成员权限保护。
- 查看与操作任务可以累计热度。
- 团队可查看热门任务排行。

### M3：通知与工作台

状态：开发中，指派通知、已读与未读数链路已验收

- 任务事件产生通知。
- 用户可查看通知并标记已读。
- Redis 快速返回未读消息数。
- 动态流展示关键协作动作。

### M4：稳定性与工程化

状态：未开始

- 关键接口具有限流。
- 主要业务场景有测试或可重复的验证脚本。
- 明确 MySQL 与 Redis 双写失败时的处理策略。
- 可继续加入锁与延迟提醒。

## 15. 每个功能的学习与开发流程

后续保持“你实现、Codex 辅导”的节奏：

```text
1. Codex 先详细说明当前小阶段的目标、实现思路、数据流、权限规则、异常处理理由与验收点
2. Codex 先提供仅保留步骤注释的空白代码框架，便于你看清实现顺序
3. 空白框架后紧接着给出带必要注释的完整参考代码，由你直接对照完成实现，不需要再次追问
4. 完整参考代码后，Codex 解释代码中容易出错、涉及数据一致性或值得重点理解的地方
5. 按 api -> logic -> controller -> 联调/测试 的顺序完成该小阶段
6. 用 curl 验证成功路径、权限边界和幂等行为
7. 观察 MySQL 记录与 Redis key/返回结果，确认缓存、状态或排行符合预期
8. 每完成并验收一个小阶段，Codex 同步完善本路线图的状态、验收记录和下一步
```

### 每日收尾约定

当你说“今天就到这里”时，Codex 自动完成当天工作的收尾：

1. 只针对当天已经实现并完成验收的功能，补齐 `resource/public` 中相应的前端展示、操作入口和交互反馈；尚未实现或未验收的接口不提前接入页面。
2. 运行与当天改动相关的格式化、编译、测试或可执行联调验证，并记录未能验证的部分。
3. 更新本路线图中的完成状态与验收结果，并新增当天完成事项记录，说明完成的接口、规则、测试和前端收尾内容。
4. 在本路线图中规划第二天任务，写明下一小阶段的目标、首个实施步骤和需要验证的关键行为。
5. 创建一次仅包含本阶段相关代码与文档文件的 Git 提交，提交范围不包含工作区中与本阶段无关的已有改动。

## 16. 当前立即下一步

任务主线验收、任务编辑测试和现有读取权限修复均已完成。下一阶段进入通知中心设计审核与实现。

### 已完成：手工验收任务工作台

已通过控制台验证：

1. 创建任务，并在列表和详情中看到一致的数据。
2. 通过 `PUT /tasks/{taskId}` 修改标题、说明、负责人和优先级。
3. 通过 `assigneeId = 0` 解除负责人。
4. 将状态从 `todo` 流转到 `doing`、`done`。
5. 确认动态流出现创建、编辑、状态变更记录，热门任务排行随查看和修改增加分数。

同时已验证非团队成员不能查询、编辑或修改任务状态，且负责人不能指派给其他团队成员。

### 已完成：收口任务逻辑测试

已通过：非团队成员编辑失败、跨团队负责人失败、无变化提交不重复增加动态和热度，以及成功更新字段并写空负责人。

写入测试在清理阶段直接恢复 MySQL 字段，并移除测试产生的动态、回退热度；连续运行两次测试均通过。

### 已完成：关闭读取权限缺口

以下读取接口已经补齐 Logic 层团队成员权限，并完成成员/非成员手工验收：

```http
GET /teams/{teamId}/members
GET /teams/{teamId}/online-members
```

### 已完成：创建通知表并生成模型

`notification` 表已创建，并已通过 `make dao` 生成 `dao`、`do`、`entity` 模型；字段与第一期设计一致。

### 已完成：任务指派通知与已读链路

已完成通知列表、标记已读和内部创建通知，并将 `task_assigned` 接入 `UpdateTask`。已验证 MySQL 通知写入、Redis DB 1 未读 Set 写入与清理、越权拒绝及重复已读幂等。

### 已完成：未读通知数量接口

已完成 `GET /notifications/unread-count`：优先读取 Redis Set 数量，集合为空时从 MySQL 未读记录重建 Set 并返回真实数量。已通过临时隔离数据验证无未读返回 `0`、Redis 命中、MySQL 回填、JWT 拦截和标记已读后的清零流程；验收后已清理临时账号、通知与 Redis key。

### 已完成：通知中心前端收尾

控制台页面已增加通知未读指标与通知中心卡片，接入通知列表、未读数量查询和标记已读操作；登录成功后自动加载当前用户通知，标记已读后同步刷新列表与未读徽标。

### 当前第一步：接入任务状态变化通知

下一步在 `UpdateStatus` 中触发 `task_status_updated`：仅当任务状态实际变化、当前负责人非零且不是操作者本人时创建通知，并继续复用已完成的列表、未读数与已读链路验收。

## 17. 每日进展记录

### 2026-05-27：今天完成了什么

- 完成成员列表与在线成员列表的读取权限收口，非团队成员不能读取目标团队信息。
- 完成任务编辑主线验收与关键自动测试，并修正测试中固定用户 ID 随团队成员变化而失真的问题，改为动态选择真实成员与非成员。
- 创建通知表及模型，完成通知列表、标记已读、内部创建通知和 `UpdateTask -> task_assigned` 联动。
- 完成 `GET /notifications/unread-count`，验证 Redis 命中、Redis 为空时从 MySQL 重建、JWT 保护与已读清零流程。
- 完成通知中心前端收尾：页面展示未读数量、通知列表与标记已读操作，并在登录后自动刷新当前用户通知。
- 验证通过：`node --check resource/public/resource/js/app.js` 与 `go test ./... -count=1`；通知接口已完成 HTTP、MySQL 和 Redis 联调。当前内置浏览器没有可用会话，页面布局与实际点击的可视化复验留到下一次启动可用浏览器后补做。

### 2026-05-28：明天任务规划

目标：完成任务状态变化通知，让负责人除了接收任务指派外，也能收到状态流转提醒。

1. 先在 `internal/logic/task/task.go` 的 `UpdateStatus` 梳理现有状态变更流程，确定通知触发点位于任务状态成功写入之后。
2. 增加 `task_status_updated` 通知类型，并仅在状态实际变化、负责人非零且负责人不是操作者本人时调用内部创建通知。
3. 验证负责人可在通知中心看到新的未读状态通知，未读数量增加且标记已读后 Redis Set 被正确清理。
4. 有可用浏览器会话时，先复验今天补齐的通知卡片，再将新的状态通知流程纳入页面联调记录。
