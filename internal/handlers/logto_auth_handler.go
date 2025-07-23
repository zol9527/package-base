// logto 用户认证模块

package handlers

import (
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

// callback 处理 Logto 回调
func (h *LogtoAuthHandler) LogtoCallback(c *fiber.Ctx) error {
	// 获取 Logto 客户端实例
	logtoClient, ok := c.Locals("logto_client").(*client.LogtoClient)
	if !ok {
		h.logger.Error("Logto client not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, ""))
	}

	// 调用 logto callback 方法
	httpRequest := &http.Request{}
	err := fasthttpadaptor.ConvertRequest(c.Context(), httpRequest, true)
	if err != nil {
		h.logger.Error("Failed to convert request", interfaces.ErrorField(err))
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Internal server error"))
	}

	err = logtoClient.HandleSignInCallback(httpRequest)
	if err != nil {
		h.logger.Error("Logto callback failed", interfaces.ErrorField(err))
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "Bad request"))
	}

	return c.Redirect("/", fiber.StatusFound)
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
	sessionStore := c.Locals("session_store").(*session.Session)
	currentSess := sessionStore.Get(c)
	if err := currentSess.Destroy(); err != nil {
		h.logger.Error("Failed to destroy session", interfaces.ErrorField(err))
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to logout, internal error occurred"))
	}

	return c.Redirect(signOutUri, fiber.StatusFound)
}

// HandleLogin 处理用户登录
func (h *LogtoAuthHandler) LogtoLogin(c *fiber.Ctx) error {
	// 获取 Logto 客户端实例
	logtoClient, ok := c.Locals("logto_client").(*client.LogtoClient)
	if !ok {
		h.logger.Error("Logto client not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Logto client not found"))
	}

	// 调用 Logto 登录方法
	authURL, err := logtoClient.SignInWithRedirectUri(h.cfg.Logto.GetCallbackURI())
	if err != nil {
		h.logger.Error("Failed to login", interfaces.ErrorField(err))
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "Failed to login"))
	}

	return c.Redirect(authURL, fiber.StatusFound)
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
