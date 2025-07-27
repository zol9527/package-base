// logto 用户认证模块

package handlers

import (
	"fmt"
	"net/http"
	"package-base/internal/core/config"
	"package-base/internal/core/interfaces"
	"package-base/internal/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/session/v2"
	"github.com/logto-io/go/v2/client"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type LogtoAuthHandler struct {
	logger interfaces.Logger
	cfg    *config.Config
}

// 初始化 logto 用户认证处理器
func NewLogtoAuthHandler(logger interfaces.Logger, cfg *config.Config) *LogtoAuthHandler {
	return &LogtoAuthHandler{
		logger: logger,
		cfg:    cfg,
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// callback 处理 Logto 回调
func (h *LogtoAuthHandler) LogtoCallback(c *fiber.Ctx) error {

	log := h.logger.Named("logto.callback")

	// 获取 client
	logtoClient := c.Locals("logto_client").(*client.LogtoClient)
	if logtoClient == nil {
		log.Error("Logto client not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Internal server error"))
	}

	// fiber request 转化为 http request
	ctx := c.Context()
	var httpReq http.Request
	if err := fasthttpadaptor.ConvertRequest(ctx, &httpReq, false); err != nil { // 添加错误检测
		h.logger.Error("ConvertRequest failed", interfaces.ErrorField(err))
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Internal server error"))
	}

	// fix: 修复 RequestURI 中不存的问题
	httpReq.RequestURI = httpReq.URL.RequestURI()

	// 使用 callback
	err := logtoClient.HandleSignInCallback(&httpReq) // 传递指针
	if err != nil {
		log.Error("Logto callback failed",
			interfaces.ErrorField(err),
			interfaces.String("error_type", fmt.Sprintf("%T", err)),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(500, "Received an error from upstream service"))
	}

	return c.Redirect("http://localhost:3000", fiber.StatusFound)
}

// LogtoLogout 处理 Logto 登出
func (h *LogtoAuthHandler) LogtoLogout(c *fiber.Ctx) error {
	// 获取 Logto 客户端实例
	logtoClient, ok := c.Locals("logto_client").(*client.LogtoClient)
	if !ok {
		h.logger.Error("Logto client not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Logto client not found"))
	}

	// 调用 Logto 登出方法
	signOutUri, signOutErr := logtoClient.SignOut("/")
	if signOutErr != nil {
		h.logger.Error("Failed to logout", interfaces.ErrorField(signOutErr))
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "Failed to logout"))
	}

	// 获取当前会话 session 并销毁
	sessionStore, ok := c.Locals("session_store").(*session.Store)
	if !ok {
		h.logger.Error("Session store not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to logout"))
	}
	if err := sessionStore.Destroy(); err != nil {
		h.logger.Error("Failed to destroy session", interfaces.ErrorField(err))
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to logout, internal error occurred"))
	}

	return c.Redirect(signOutUri, fiber.StatusFound)
}

// HandleLogin 处理用户登录
func (h *LogtoAuthHandler) LogtoLogin(c *fiber.Ctx) error {
	log := h.logger.Named("logto.login")

	// 获取 Logto 客户端实例
	logtoClient, ok := c.Locals("logto_client").(*client.LogtoClient)
	if !ok {
		log.Error("Logto client not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Internal server error"))
	}

	// 调用 Logto 登录方法
	authURL, err := logtoClient.SignInWithRedirectUri(h.cfg.Logto.GetCallbackURI())
	if err != nil {
		log.Error("Logto登录失败",
			interfaces.ErrorField(err),
			interfaces.String("error_type", fmt.Sprintf("%T", err)),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(500, "Internal server error"))
	}

	return c.Redirect(authURL, fiber.StatusFound)
}

// LogtoValidateToken 使用 jwts 验证 token
func (h *LogtoAuthHandler) LogtoValidateToken(c *fiber.Ctx) error {
	return nil
}

// 注册路由组
func (h *LogtoAuthHandler) RegisterLogtoAuthRoutes(group fiber.Router) {

	// Logto 登录路由
	group.Get("/login", h.LogtoLogin)

	// Logto 回调路由
	group.Get("/callback", h.LogtoCallback)

	// Logto 登出路由
	group.Get("/logout", h.LogtoLogout)
}
