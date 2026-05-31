package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	defaultSecret = "redis-demo-jwt-secret" // JWT 的默认签名密钥，实际使用中应该从配置文件或环境变量中获取，并且要足够复杂以保证安全性。
	defaultIssuer = "redis-demo"            // JWT 的默认发行者，可以根据实际情况修改。
	defaultExpire = 2 * time.Hour
)

type UserClaims struct {
	UserId                  int64  `json:"userId"`
	Username                string `json:"username"`
	jwtlib.RegisteredClaims        // RegisteredClaims 包含了 JWT 标准字段，如 exp、iat、nbf 等。
}

func GenerateToken(ctx context.Context, userId int64, username string) (string, error) {
	now := time.Now()
	claims := UserClaims{
		UserId:   userId,
		Username: username,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    getIssuer(ctx),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(getExpire())),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims) // 使用 HMAC SHA256 签名算法，实际使用中可以根据需要选择其他算法。
	return token.SignedString([]byte(getSecret(ctx)))                // 使用配置中的 secret 作为签名密钥，生成 JWT 字符串。
}

func ParseToken(ctx context.Context, tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (interface{}, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, errors.New("unexpected jwt signing method")
		}
		return []byte(getSecret(ctx)), nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, errors.New("invalid jwt token")
	}
	return claims, nil
}

func getSecret(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "jwt.secret", defaultSecret).String() // 获取配置文件中的secret，如果没有则使用默认值
}

func getIssuer(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "jwt.issuer", defaultIssuer).String()
}

func getExpire() time.Duration {
	return defaultExpire
}
