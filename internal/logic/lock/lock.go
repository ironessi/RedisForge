package lock

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

// Lock 表示一次成功获取到的 Redis 分布式锁。
// Key 是锁的 Redis key，Token 是当前请求持有这把锁的凭证。
type Lock struct {
	Key   string
	Token string
}

// TryLock 尝试获取 Redis 分布式锁。
// locked=false 表示锁已经被其他请求持有，不代表系统错误。
func TryLock(ctx context.Context, key string, ttl time.Duration) (*Lock, bool, error) {
	// 1. 生成随机 token
	token, err := randomToken()
	if err != nil {
		return nil, false, err
	}
	// 2. SET key token NX EX ttl
	// 2. SET key token NX EX seconds
	// NX：只有 key 不存在时才设置成功，保证同一时间只有一个请求拿到锁。
	// EX：设置锁过期时间，避免持锁请求崩溃后锁永远不释放。
	result, err := g.Redis().Do(ctx, "SET", key, token, "NX", "EX", int64(ttl.Seconds()))
	if err != nil {
		return nil, false, err
	}
	// 3. Redis 返回 OK，说明拿锁成功
	if result.String() == "OK" {
		return &Lock{
			Key:   key,
			Token: token,
		}, true, nil
	}
	// 4. Redis 没返回 OK，说明锁被别人持有
	return nil, false, nil
}

func Unlock(ctx context.Context, lock *Lock) error {
	// 1. lock == nil 直接返回
	if lock == nil {
		return nil
	}
	// 2. 调用 UnlockWithToken
	err := UnlockWithToken(ctx, lock.Key, lock.Token)
	return err
}

// UnlockWithToken 使用 Lua 脚本释放锁。
// Lua 可以保证“读取 value、比较 token、删除 key”在 Redis 内原子执行。
func UnlockWithToken(ctx context.Context, key string, token string) error {
	// 1. 执行 Lua 脚本
	_, err := g.Redis().Do(ctx, "EVAL", unlockScript, 1, key, token)
	return err
}

// randomToken 生成锁 token。
// 这里使用 crypto/rand，而不是 math/rand，因为锁 token 要尽量避免被猜到或重复。
func randomToken() (string, error) {
	// 1. 生成 16 字节随机数
	bytes := make([]byte, 16)
	if _, err := cryptoRand.Read(bytes); err != nil {
		return "", err
	}
	// 2. 转成十六进制字符串
	return hex.EncodeToString(bytes), nil
}
