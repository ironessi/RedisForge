package lock

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func TestTryLockAndUnlock(t *testing.T) {
	ctx := gctx.New()
	key := fmt.Sprintf("lock:test:%d", time.Now().UnixNano())
	ttlSeconds := int64(10)

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean test lock failed: %v", err)
		}
	})

	lock, locked, err := TryLock(ctx, key, time.Duration(ttlSeconds)*time.Second)
	if err != nil {
		t.Fatalf("try lock failed: %v", err)
	}
	if !locked {
		t.Fatal("first try lock should succeed")
	}
	if lock == nil {
		t.Fatal("lock should not be nil after successful try lock")
	}
	if lock.Key != key {
		t.Fatalf("unexpected lock key: got=%q want=%q", lock.Key, key)
	}
	if lock.Token == "" {
		t.Fatal("lock token should not be empty")
	}

	value, err := g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read lock value failed: %v", err)
	}
	if value.String() != lock.Token {
		t.Fatalf("redis lock value should equal token: got=%q want=%q", value.String(), lock.Token)
	}

	ttl, err := g.Redis().TTL(ctx, key)
	if err != nil {
		t.Fatalf("read lock ttl failed: %v", err)
	}
	if ttl <= 0 || ttl > ttlSeconds {
		t.Fatalf("lock ttl should be between 1 and %d seconds, got %d", ttlSeconds, ttl)
	}

	secondLock, secondLocked, err := TryLock(ctx, key, time.Duration(ttlSeconds)*time.Second)
	if err != nil {
		t.Fatalf("second try lock failed: %v", err)
	}
	if secondLocked {
		t.Fatalf("second try lock should fail while lock is held: %+v", secondLock)
	}

	if err := UnlockWithToken(ctx, key, "wrong-token"); err != nil {
		t.Fatalf("unlock with wrong token should not return error: %v", err)
	}
	value, err = g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read lock value after wrong unlock failed: %v", err)
	}
	if value.IsNil() || value.String() != lock.Token {
		t.Fatalf("wrong token should not delete lock, got=%q", value.String())
	}

	if err := Unlock(ctx, lock); err != nil {
		t.Fatalf("unlock with correct token failed: %v", err)
	}
	value, err = g.Redis().Get(ctx, key)
	if err != nil {
		t.Fatalf("read lock value after unlock failed: %v", err)
	}
	if !value.IsNil() {
		t.Fatalf("unlock should delete lock, got=%q", value.String())
	}
}

func TestUnlockNilLock(t *testing.T) {
	ctx := gctx.New()

	if err := Unlock(ctx, nil); err != nil {
		t.Fatalf("unlock nil lock should return nil, got %v", err)
	}
}
