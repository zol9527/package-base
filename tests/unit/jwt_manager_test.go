package unit

import (
	"package-base/internal/core/utils"
	"testing"
	"time"
)

// TestJWTManager_AccessTokenUniqueness 同一秒内多次生成 access_token 仍应不同（因 JTI 引入）
func TestJWTManager_AccessTokenUniqueness(t *testing.T) {
	jm := utils.NewJWTManager("secret", 5*time.Minute, 30*time.Minute)
	tok1, _, err := jm.GenerateAccessToken("u1", "tester", "t@example.com", 0)
	if err != nil {
		t.Fatalf("生成第一个 token 失败: %v", err)
	}
	// 不 sleep，确保同一秒
	tok2, _, err := jm.GenerateAccessToken("u1", "tester", "t@example.com", 0)
	if err != nil {
		t.Fatalf("生成第二个 token 失败: %v", err)
	}
	if tok1 == tok2 {
		t.Fatalf("期望同一秒生成的两个 access token 不同, 实际相同: %s", tok1)
	}
}
