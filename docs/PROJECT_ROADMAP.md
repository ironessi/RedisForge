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
| 第一阶段 | 用户与团队 | 基础功能已完成，部分读取权限待收紧 |
| 第二阶段 | 在线状态 | 基础功能已完成，在线成员读取权限待收紧 |
| 第三阶段 | 任务/工单流转 | 开发中，任务基本编辑接口与控制台工作台已接入，待联调验收 |
| 第四阶段 | 通知中心与未读数 | 未开始 |
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
| 通知未读集合 | `notification:unread:{userId}` | Set | 可由 MySQL 重建 | 待开发 |
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

当前 `GET /teams/{teamId}/members` 仅要求登录。后续应改成只有该团队成员才可以查看。

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

`GET /teams/{teamId}/online-members` 当前仅检查已登录，应增加“查询者必须属于该团队”的 Logic 权限校验。

## 7. 第三阶段：任务/工单流转

### 目标

完成协作系统的核心实体：团队成员可以创建任务、查看任务、推动任务状态变化，并为动态流、通知和排行提供业务事件。

### 当前已完成

| Method | Path | 用途 | 状态 |
| --- | --- | --- | --- |
| `POST` | `/teams/{teamId}/tasks` | 创建任务 | 已完成并手工验证 |
| `GET` | `/teams/{teamId}/tasks` | 查询团队任务列表 | 已完成并手工验证 |
| `GET` | `/tasks/{taskId}` | 查询任务详情并累计浏览热度 | 已完成并手工验证 |
| `PUT` | `/tasks/{taskId}` | 更新任务信息 | 已实现，待手工验证 |
| `PATCH` | `/tasks/{taskId}/status` | 更新任务状态 | 状态流转已验证，热度计分待复验 |
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

### 主线接口待验收

| Method | Path | 用途 | 验收要点 |
| --- | --- | --- | --- |
| `PUT` | `/tasks/{taskId}` | 更新任务信息 | 校验团队权限、负责人归属和解除分配 |

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
- 编辑成功后写入 `task_updated` 团队动态并增加一次排行热度；手工联调验收待执行。
- 控制台已加入任务创建、列表、详情编辑与状态切换入口，并在操作后刷新动态流和热门排行。

## 8. 第四阶段：通知中心与已读未读

### 目标

将任务流转产生的事件送达到相关员工，例如任务被分配、状态变化、有人提及自己等。

### 计划接口

| Method | Path | 用途 |
| --- | --- | --- |
| `POST` | `/notifications` | 创建通知 |
| `GET` | `/notifications` | 查询当前用户通知列表 |
| `POST` | `/notifications/{id}/read` | 标记单条通知已读 |
| `GET` | `/notifications/unread-count` | 查询当前用户未读数 |

### 数据设计

MySQL 保存通知主数据，至少包含：

```text
id
receiver_id
type
content
related_task_id
is_read
created_at
read_at
```

Redis 使用 Set 加速未读数量：

```redis
SADD  notification:unread:1 1001
SREM  notification:unread:1 1001
SCARD notification:unread:1
```

### 业务流程

创建通知：

```text
插入 MySQL notification
-> SADD notification:unread:{receiverId} notificationId
```

标记已读：

```text
更新 MySQL is_read/read_at
-> SREM notification:unread:{receiverId} notificationId
```

查询未读数：

```text
SCARD notification:unread:{userId}
```

### 与任务模块的联动

通知中心完成后，应从以下任务事件自动产生通知：

- 任务被分配给某位成员
- 任务状态被变更
- 后续增加评论或提及时通知相关人员

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
| 1 | 验收任务更新 `PUT /tasks/{taskId}` | 验证任务流转主线 |
| 2 | 实现通知中心与未读数 | 让任务事件真正触达员工 |
| 3 | 收紧成员列表、在线成员读取权限 | 消除现有数据泄露风险 |
| 4 | 完善动态事件与手工验收 | 让工作台信息流完整 |
| 5 | 接口限流 | 为关键写接口和登录增加保护 |
| 6 | 自动测试、README 与演示脚本 | 形成可展示项目 |
| 7 | 锁与延迟任务 | 作为进阶能力 |

## 14. 里程碑与验收标准

### M1：团队协作基础

状态：接近完成

- 员工可注册、登录、退出。
- 可创建团队并添加成员。
- 用户资料与团队成员使用 Redis 缓存。
- 在线心跳和动态查询基本可用。
- 补齐成员、在线成员读取权限后完成 M1。

### M2：任务流转与热度

状态：进行中

- 创建、列表、详情、状态流转、热门排行和基本编辑均已实现；基本编辑待联调验收。
- 所有任务读取和修改都有团队成员权限保护。
- 查看与操作任务可以累计热度。
- 团队可查看热门任务排行。

### M3：通知与工作台

状态：未开始

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
1. 先明确接口与业务规则
2. 在 api 层定义请求、响应和校验
3. 在 logic 层实现存在性检查、权限、MySQL / Redis 行为
4. 在 controller 层从 JWT 获取当前用户并调用 logic
5. 编译测试
6. 用 curl 验证成功路径与越权路径
7. 观察 Redis key 或返回结果，确认缓存/状态/排行符合预期
```

## 16. 当前立即下一步

当前不继续扩张范围，先完成任务编辑能力的手工验收：

```http
PUT /tasks/{taskId}
```

验证成员可编辑、非成员不可编辑、跨团队负责人不可指派，以及 `assigneeId = 0` 可解除负责人。
