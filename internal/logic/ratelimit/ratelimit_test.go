package ratelimit

import (
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func TestCheckTaskCreateLimitsAfterTenRequests(t *testing.T) {
	ctx := gctx.New()
	userId := uint64(time.Now().UnixNano())
	key := taskCreateRateKey(userId, time.Now().Unix()/60)

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean task create rate key failed: %v", err)
		}
	})

	for i := 1; i <= taskCreateLimit; i++ {
		if err := CheckTaskCreate(ctx, userId); err != nil {
			t.Fatalf("request %d should be allowed: %v", i, err)
		}
	}

	err := CheckTaskCreate(ctx, userId)
	if err == nil {
		t.Fatal("request over task create limit should be rejected")
	}
	if !strings.Contains(err.Error(), "请求过于频繁") {
		t.Fatalf("unexpected task create limit error: %v", err)
	}

	ttl, err := g.Redis().TTL(ctx, key)
	if err != nil {
		t.Fatalf("read task create rate key ttl failed: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("task create rate key should have ttl, got %d", ttl)
	}
}

func TestCheckLoginLimitsAfterFiveRequests(t *testing.T) {
	ctx := gctx.New()
	ip := fmt.Sprintf("test-login-%d", time.Now().UnixNano())
	key := loginRateKey(ip, time.Now().Unix()/60)

	t.Cleanup(func() {
		if _, err := g.Redis().Del(ctx, key); err != nil {
			t.Errorf("clean login rate key failed: %v", err)
		}
	})

	for i := 1; i <= loginLimit; i++ {
		if err := CheckLogin(ctx, ip); err != nil {
			t.Fatalf("login request %d should be allowed: %v", i, err)
		}
	}

	err := CheckLogin(ctx, ip)
	if err == nil {
		t.Fatal("request over login limit should be rejected")
	}
	if !strings.Contains(err.Error(), "登录过于频繁") {
		t.Fatalf("unexpected login limit error: %v", err)
	}

	ttl, err := g.Redis().TTL(ctx, key)
	if err != nil {
		t.Fatalf("read login rate key ttl failed: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("login rate key should have ttl, got %d", ttl)
	}
}
