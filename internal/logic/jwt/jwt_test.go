package jwt

import (
	"testing"

	"github.com/gogf/gf/v2/os/gctx"
)

func TestGenerateAndParseToken(t *testing.T) {
	ctx := gctx.New()

	token, err := GenerateToken(ctx, 1, "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	claims, err := ParseToken(ctx, token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserId != 1 {
		t.Fatalf("UserId = %d, want 1", claims.UserId)
	}
	if claims.Username != "admin" {
		t.Fatalf("Username = %q, want admin", claims.Username)
	}
}
