# RedisForge

RedisForge is a GoFrame + Redis backend lab for JWT auth, token blacklist logout, captcha TTL, user caching, and cache reliability patterns.

## Features

- User registration with bcrypt password hashing
- JWT login and authentication middleware
- Redis token blacklist for logout
- Redis captcha storage with TTL
- MySQL access through GoFrame DAO

## Stack

- GoFrame
- MySQL
- Redis
- JWT

## Quick Start

1. Copy the example configs and update local credentials.

```bash
cp manifest/config/config.example.yaml manifest/config/config.yaml
cp hack/config.example.yaml hack/config.yaml
```

2. Start MySQL and Redis, then run the service.

```bash
go run main.go
```

3. Try the auth APIs.

```bash
curl -X POST http://127.0.0.1:8000/auth/captcha \
  -H "Content-Type: application/json" \
  -d '{"username":"test001"}'
```

## Notes

Local config files are ignored by git. Keep real database passwords and JWT secrets out of the repository.
