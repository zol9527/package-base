package integration

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test_NoLogtoDependencies 确认代码库中已无 logto 相关引用 (US1 T020)
func Test_NoLogtoDependencies(t *testing.T) {
	// 在项目根执行 grep -R logto ，忽略 tasks 与 markdown 的历史描述
	// 仅扫描核心源码目录，避免文档/元数据触发
	grepScript := "(grep -R 'logto' -n cmd internal go.mod go.sum || true) | grep -v tests/integration/startup_nologto_test.go"
	cmd := exec.Command("bash", "-c", grepScript)
	cmd.Dir = rootDir(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// grep 失败以外的错误
		if exitErr, ok := err.(*exec.ExitError); ok {
			// 忽略无匹配的退出码 1，通过输出判断
			_ = exitErr
		} else {
			t.Fatalf("执行 grep 失败: %v", err)
		}
	}
	output := string(out)
	// 允许 tasks.md 里出现规划性文字，其余不应出现
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.Contains(line, "specs/001-logto-session/tasks.md") {
			continue
		}
		// 若剩余任何行则失败
		t.Fatalf("发现残余 logto 引用: %s", line)
	}
}

// rootDir 推断项目根目录
func rootDir(t *testing.T) string {
	dir, err := filepath.Abs("./")
	if err != nil {
		t.Fatalf("获取目录失败: %v", err)
	}
	return dir + "/../../" // tests/integration -> project root
}
