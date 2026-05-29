# TeamPulse

TeamPulse 是一个基于 GoFrame、MySQL 和 Redis 的团队协作后端练习项目。

它模拟公司内部协作系统中的常见能力：员工注册登录、团队成员管理、在线状态、任务流转、通知中心、团队动态流、热门任务排行和关键接口限流。项目重点不是把 Redis 当主数据库，而是练习 Redis 在真实业务里承担的缓存、TTL 状态、集合、排行榜、计数器和消息辅助能力。

## 功能概览

- 用户注册、登录、JWT 鉴权和退出登录
- 登录验证码 TTL 存储
- 用户资料缓存
- 团队创建、添加成员和成员缓存
- 团队在线成员心跳与查询
- 任务创建、查询、编辑、状态流转和热门排行
- 任务指派通知与任务状态变化通知
- 通知列表、未读数和标记已读
- 团队操作动态流
- 创建任务限流和登录限流
- 控制台前端页面，位于 `resource/public`

## Redis 使用点

| 能力 | Key | 类型 | 说明 |
| --- | --- | --- | --- |
| 登录验证码 | `auth:captcha:{username}` | String | 5 分钟 TTL |
| JWT 退出黑名单 | `jwt:blacklist:{token}` | String | TTL 与 token 有效期配合 |
| 用户资料缓存 | `user:profile:{userId}` | String(JSON) | Cache Aside |
| 团队成员缓存 | `team:members:{teamId}` | Set | 缓存团队成员 ID |
| 用户在线状态 | `presence:user:{userId}` | String | 60 秒心跳 TTL |
| 团队在线候选成员 | `presence:team:{teamId}` | Set | 查询在线成员时过滤失效心跳 |
| 团队动态流 | `team:activities:{teamId}` | List | 最近团队操作 |
| 热门任务排行 | `team:task:hot:{teamId}` | Sorted Set | 任务浏览和操作热度 |
| 通知未读集合 | `notification:unread:{userId}` | Set | 成员是 notificationId |
| 创建任务限流 | `rate:task:create:{userId}:{minute}` | String Counter | 每用户每分钟最多 10 次 |
| 登录限流 | `rate:login:{ip}:{minute}` | String Counter | 每 IP 每分钟最多 5 次 |

## 技术栈

- GoFrame v2
- MySQL
- Redis
- JWT
- bcrypt
- GoFrame DAO

## 本地启动

1. 复制配置文件。

```bash
cp manifest/config/config.example.yaml manifest/config/config.yaml
cp hack/config.example.yaml hack/config.yaml
```

2. 修改本地 MySQL、Redis 和 JWT 配置。

```yaml
database:
  default:
    link: "mysql:root:password@tcp(127.0.0.1:3306)/redis_forge"

redis:
  default:
    address: 127.0.0.1:6379
    db: 1

jwt:
  secret: "change-me-in-local-config"
  issuer: "redis-forge"
```

3. 启动服务。

```bash
go run main.go
```

默认服务地址：

```text
http://127.0.0.1:8000
```

OpenAPI 文档：

```text
http://127.0.0.1:8000/swagger
```

控制台页面：

```text
http://127.0.0.1:8000/
```

## 测试

运行全部测试：

```bash
go test ./... -count=1
```

当前重点测试覆盖：

- JWT 生成与解析
- 任务编辑权限和副作用
- 任务状态通知创建规则
- 创建任务限流
- 登录限流

部分测试会访问本地 MySQL 和 Redis，并在结束时清理测试产生的 Redis key、通知、动态或任务副作用。

## API 快速演示

下面命令展示主流程。实际账号、验证码、团队 ID 和任务 ID 请以本地数据为准。

### 获取验证码

```bash
curl -s -X POST 'http://127.0.0.1:8000/auth/captcha' \
  -H 'Content-Type: application/json' \
  -d '{"username":"test001"}'
```

### 登录并保存 Token

```bash
TOKEN=$(curl -s -X POST 'http://127.0.0.1:8000/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"test001","password":"123456","captcha":"验证码"}' \
  | jq -r '.data.token')
```

### 创建团队

```bash
curl -s -X POST 'http://127.0.0.1:8000/teams' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"name":"研发协作组"}'
```

### 添加团队成员

```bash
curl -s -X POST 'http://127.0.0.1:8000/teams/1/members' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"userId":2,"role":"member"}'
```

### 上报在线心跳

```bash
curl -s -X POST 'http://127.0.0.1:8000/presence/heartbeat' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"teamId":1}'
```

### 创建任务

```bash
curl -s -X POST 'http://127.0.0.1:8000/teams/1/tasks' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"title":"接入通知中心","description":"完成任务通知链路","assigneeId":2,"priority":2}'
```

### 更新任务状态

```bash
curl -s -X PATCH 'http://127.0.0.1:8000/tasks/1/status' \
  -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"status":"doing"}'
```

### 查询通知和未读数

```bash
curl -s 'http://127.0.0.1:8000/notifications' \
  -H "Authorization: Bearer ${TOKEN}"

curl -s 'http://127.0.0.1:8000/notifications/unread-count' \
  -H "Authorization: Bearer ${TOKEN}"
```

### 标记通知已读

```bash
curl -s -X PATCH 'http://127.0.0.1:8000/notifications/1/read' \
  -H "Authorization: Bearer ${TOKEN}"
```

### 查询团队动态和热门任务

```bash
curl -s 'http://127.0.0.1:8000/teams/1/activities' \
  -H "Authorization: Bearer ${TOKEN}"

curl -s 'http://127.0.0.1:8000/teams/1/tasks/hot' \
  -H "Authorization: Bearer ${TOKEN}"
```

## 限流验证

创建任务限流：

```text
同一用户一分钟内前 10 次创建任务成功，第 11 次返回“请求过于频繁，请稍后再试”。
```

登录限流：

```text
同一 IP 一分钟内前 5 次登录请求继续进入原登录逻辑，第 6 次返回“登录过于频繁，请稍后再试”。
```

## 开发路线

完整路线和每日进展记录见：

```text
docs/PROJECT_ROADMAP.md
```

## 安全说明

- `manifest/config/config.yaml` 和 `hack/config.yaml` 是本地配置文件，不应提交真实密码。
- 示例验证码接口会直接返回验证码，只用于学习和本地测试。
- JWT secret 请在本地配置中替换，不要使用示例值部署。
