package testhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"package-base/internal/container"
	"package-base/internal/core/config"
	"package-base/internal/core/log"
	"package-base/internal/routes"

	"github.com/gofiber/fiber/v2"
)

// Function: StartTestApp
//
// Description: 启动测试用 Fiber 应用，自动：加载配置、创建 logger、初始化容器、调用路由装配、监听并等待端口就绪；返回关闭函数用于清理。
//
// Parameters:
//
//	cfgPath(string): 配置文件路径
//	addr(string): 监听地址 host:port
//	disableConsole(bool): 是否关闭控制台日志输出
//	waitTimeout(time.Duration): 等待端口就绪最大时长
//
// Returns:
//
//	app(*fiber.App): Fiber 应用实例
//	cont(*container.Container): 容器实例
//	shutdown(func()): 清理函数 (关闭 Fiber 与 容器)
//	err(error): 可能的错误
func StartTestApp(cfgPath, addr string, disableConsole bool, waitTimeout time.Duration) (*fiber.App, *container.Container, func(), error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("加载配置失败: %w", err)
	}

	loggerCfg := &log.LoggerConfig{Level: cfg.Logger.Level, Format: cfg.Logger.Format, EnableConsole: !disableConsole, EnableFile: false}
	logger, err := log.NewLogger(loggerCfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("创建logger失败: %w", err)
	}

	cont := container.NewContainer(cfg, logger)
	if err := cont.Initialize(context.Background()); err != nil {
		return nil, nil, nil, fmt.Errorf("容器初始化失败: %w", err)
	}

	app := fiber.New()
	routes.NewRoutes(app, cont).Setup()

	go app.Listen(addr)
	if ok, werr := WaitForServer(addr, waitTimeout, 50*time.Millisecond); !ok {
		cont.Close()
		_ = app.Shutdown()
		return nil, nil, nil, fmt.Errorf("服务器未在超时内启动: %v", werr)
	}

	shutdown := func() {
		_ = app.Shutdown()
		cont.Close()
	}
	return app, cont, shutdown, nil
}

// Function: GetJSON 执行简单 JSON POST/GET 请求（无需在多个测试重复编写 http.Client）。
//
// Description: 发送 HTTP 请求并解析通用响应结构 {code,msg,data}，data 原样保持为字节以供二次解析。
//
// Parameters:
//
//	method(string): HTTP 方法
//	url(string): 完整 URL
//	body(io.Reader): 请求体，可为 nil
//	timeout(time.Duration): 请求超时
//
// Returns:
//
//	status(int): HTTP 状态码
//	code(int): 响应业务 code
//	msg(string): 响应消息
//	data([]byte): data 字段原始 JSON
//	err(error): 可能的错误
func GetJSON(method, url string, body []byte, timeout time.Duration) (int, int, string, []byte, error) {
	client := &http.Client{Timeout: timeout}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return 0, 0, "", nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, "", nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, 0, "", nil, err
	}
	var parsed struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return resp.StatusCode, 0, "", raw, err
	}
	return resp.StatusCode, parsed.Code, parsed.Msg, parsed.Data, nil
}
