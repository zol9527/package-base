package integration

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"package-base/internal/models"
	"package-base/tests/testhelpers"
)

// Test_AuthRefreshToken_Concurrent 并发刷新同一 refresh_token 仅允许一次成功
func Test_AuthRefreshToken_Concurrent(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18092"

	_, cont, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("启动测试应用失败: %v", err)
	}
	defer shutdown()

	if err := cont.GetMySQLStorage().Migrate(t.Context(), &models.User{}); err != nil {
		t.Fatalf("用户模型迁移失败: %v", err)
	}
	if err := testhelpers.EnsureSessionTable(cont); err != nil {
		t.Fatalf("创建会话表失败: %v", err)
	}

	password := "Passw0rd!"
	if err := testhelpers.SeedTestUser(t, cont, password); err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 登录获取初始 refresh
	_, refresh, err := testhelpers.LoginForTokens(testhelpers.GetJSON, addr, "tester", password, 3*time.Second)
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}

	// 并发刷新
	const workers = 6
	var success int64
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, rErr := testhelpers.RefreshTokens(testhelpers.GetJSON, addr, refresh, 3*time.Second)
			if rErr == nil { // 认为成功
				// 竞态不影响正确性：只统计成功次数
				// 使用互斥或原子，这里为简洁采用互斥
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Fatalf("并发刷新应仅成功1次 实际 %d", success)
	}
}

var mu sync.Mutex
