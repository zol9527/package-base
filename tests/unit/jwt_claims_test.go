package unit

import (
	"testing"
	"time"

	"package-base/internal/core/utils"
)

// TestJWTClaimsPrecision 验证 JWT exp-iat 与配置的 access TTL 接近
func TestJWTClaimsPrecision(t *testing.T) {
	ttl := 3 * time.Second
	mgr := utils.NewJWTManager("secret", ttl, 10*time.Second)

	token, _, err := mgr.GenerateAccessToken("u1", "tester", "t@example.com", 1)
	if err != nil || token == "" {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := mgr.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	if claims.RegisteredClaims.ExpiresAt == nil || claims.RegisteredClaims.IssuedAt == nil {
		t.Fatalf("claims missing exp/iat")
	}
	expUnix := claims.RegisteredClaims.ExpiresAt.Time
	iatUnix := claims.RegisteredClaims.IssuedAt.Time

	delta := expUnix.Sub(iatUnix)
	if delta < ttl-500*time.Millisecond || delta > ttl+500*time.Millisecond {
		t.Fatalf("exp-iat not within tolerance: got %v want ~%v", delta, ttl)
	}

	// refresh token 精度
	rt, rExp, err := mgr.GenerateRefreshToken("u1")
	if err != nil || rt == "" {
		t.Fatalf("unexpected refresh error: %v", err)
	}
	if time.Until(rExp) <= 0 {
		t.Fatalf("refresh exp should be in the future")
	}

	// 同秒生成的两个 access token 应该不同（JTI）
	token2, _, err := mgr.GenerateAccessToken("u1", "tester", "t@example.com", 1)
	if err != nil || token2 == "" {
		t.Fatalf("unexpected error gen2: %v", err)
	}
	if token2 == token {
		t.Fatalf("two access tokens should differ due to JTI")
	}
}
