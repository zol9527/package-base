package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"package-base/tests/testhelpers"
)

// Test_HealthEndpoints /health 与 /ping 正常返回
func Test_HealthEndpoints(t *testing.T) {
	wd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(wd, "../../"))
	cfgPath := filepath.Join(root, "config.json")
	addr := "127.0.0.1:18089"

	_, _, shutdown, err := testhelpers.StartTestApp(cfgPath, addr, true, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("启动测试应用失败: %v", err)
	}
	defer shutdown()

	// /ping 直接使用 http 客户端（该端点未包装统一 {code,msg,data}）
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + addr + "/ping")
	if err != nil {
		t.Fatalf("请求 /ping 失败: %v", err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("/ping 期望 200 实际 %d", resp.StatusCode)
	}
	var pingParsed map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &pingParsed)
	if v, ok := pingParsed["message"].(string); !ok || v != "pong" {
		t.Fatalf("/ping 返回不正确: %v, 原始=%s", pingParsed, string(bodyBytes))
	}

	// /health 同样直接 http 请求（返回简单 JSON）
	hResp, hErr := client.Get("http://" + addr + "/health")
	if hErr != nil {
		t.Fatalf("请求 /health 失败: %v", hErr)
	}
	hBody, _ := io.ReadAll(hResp.Body)
	hResp.Body.Close()
	if hResp.StatusCode != 200 {
		t.Fatalf("/health 期望 200 实际 %d", hResp.StatusCode)
	}
	var healthObj map[string]interface{}
	if err := json.Unmarshal(hBody, &healthObj); err != nil {
		t.Fatalf("解析 /health 响应失败: %v 原始=%s", err, string(hBody))
	}
	if healthObj["status"] != "healthy" {
		t.Fatalf("/health status 非 healthy: %v", healthObj)
	}
	if _, ok := healthObj["app"].(string); !ok {
		t.Fatalf("/health 缺少 app 字段: %v", healthObj)
	}
	if _, ok := healthObj["version"].(string); !ok {
		t.Fatalf("/health 缺少 version 字段: %v", healthObj)
	}
}
