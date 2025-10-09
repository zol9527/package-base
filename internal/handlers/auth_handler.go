package handlers

import (
	"time"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	"package-base/internal/dto"
	"package-base/internal/middleware"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService serviceInterfaces.AuthService
	validator   *utils.Validator
	logger      interfaces.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(
	authService serviceInterfaces.AuthService,
	logger interfaces.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   utils.NewValidator(),
		logger:      logger,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} dto.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "AUTH_INVALID_REQUEST", "invalid request body", nil)
	}
	if err := h.validator.ValidateStruct(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "AUTH_VALIDATION_FAILED", err.Error(), nil)
	}
	response, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		code := "AUTH_REGISTER_FAILED"
		switch err.Error() {
		case "username already exists":
			statusCode = fiber.StatusConflict
			code = "AUTH_USERNAME_EXISTS"
		case "email already exists":
			statusCode = fiber.StatusConflict
			code = "AUTH_EMAIL_EXISTS"
		case "validation failed":
			statusCode = fiber.StatusBadRequest
			code = "AUTH_VALIDATION_FAILED"
		}
		return utils.Error(c, statusCode, code, err.Error(), nil)
	}
	return utils.Success(c, response, "registration successful")
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户账户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} dto.APIResponse{data=dto.AuthResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "AUTH_INVALID_REQUEST", "invalid request body", nil)
	}
	if err := h.validator.ValidateStruct(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "AUTH_VALIDATION_FAILED", err.Error(), nil)
	}
	response, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		code := "AUTH_LOGIN_FAILED"
		switch err.Error() {
		case "invalid credentials":
			statusCode = fiber.StatusUnauthorized
			code = "AUTH_INVALID_CREDENTIALS"
		case "account is not active":
			statusCode = fiber.StatusUnauthorized
			code = "AUTH_ACCOUNT_INACTIVE"
		case "validation failed":
			statusCode = fiber.StatusBadRequest
			code = "AUTH_VALIDATION_FAILED"
		}
		return utils.Error(c, statusCode, code, "login failed", nil)
	}
	return utils.Success(c, response, "login successful")
}

// RefreshToken 刷新Token
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} dto.APIResponse{data=dto.TokenResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid request body"))
	}

	// 验证请求参数
	if err := h.validator.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, err.Error()))
	}

	// 调用服务层刷新Token
	response, err := h.authService.RefreshToken(c.Context(), &req)
	if err != nil {
		h.logger.Warn("token refresh failed",
			interfaces.ErrorField(err),
			interfaces.String("ip", c.IP()),
		)

		statusCode := fiber.StatusInternalServerError
		message := "token refresh failed"

		errMsg := err.Error()
		switch {
		case errMsg == "invalid refresh token" || errMsg == "session expired":
			statusCode = fiber.StatusUnauthorized
			message = errMsg
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	return c.JSON(dto.SuccessWithMessage("token refreshed successfully", response))
}

// Logout 用户登出
// @Summary 用户登出
// @Description 登出当前设备
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Query("session_id", "")

	// 如果没有提供session_id，这里可以根据JWT中的信息来确定session
	// 为简化实现，假设前端会提供session_id

	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "session_id is required"))
	}

	if err := h.authService.Logout(c.Context(), userID, sessionID); err != nil {
		h.logger.Error("logout failed",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.String("session_id", sessionID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "logout failed"

		if err.Error() == "session not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	h.logger.Info("user logged out successfully",
		interfaces.String("user_id", userID),
		interfaces.String("session_id", sessionID),
	)

	return c.JSON(dto.Success("logout successful"))
}

// LogoutAll 登出所有设备
// @Summary 登出所有设备
// @Description 登出用户的所有设备会话
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	if err := h.authService.LogoutAll(c.Context(), userID); err != nil {
		h.logger.Error("logout all failed",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "logout all failed"))
	}

	h.logger.Info("user logged out from all devices",
		interfaces.String("user_id", userID),
	)

	return c.JSON(dto.Success("logged out from all devices successfully"))
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 用户修改登录密码
// @Tags 认证
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "invalid request body"))
	}

	// 验证请求参数
	if err := h.validator.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, err.Error()))
	}

	// 调用服务层修改密码
	if err := h.authService.ChangePassword(c.Context(), userID, &req); err != nil {
		h.logger.Error("change password failed",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "change password failed"

		if err.Error() == "invalid old password" {
			statusCode = fiber.StatusBadRequest
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	h.logger.Info("password changed successfully",
		interfaces.String("user_id", userID),
	)

	return c.JSON(dto.Success("password changed successfully"))
}

// GetProfile 获取当前用户信息
// @Summary 获取用户资料
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.UserInfo}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	userInfo, err := h.authService.GetCurrentUserInfo(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user profile",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to get user profile"

		if err.Error() == "user not found" {
			statusCode = fiber.StatusNotFound
			message = err.Error()
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	return c.JSON(dto.Success(userInfo))
}

// GetSessions 获取用户会话列表
// @Summary 获取用户会话
// @Description 获取当前用户的所有设备会话
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.SessionListResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/sessions [get]
func (h *AuthHandler) GetSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	sessions, err := h.authService.GetUserSessions(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get user sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get sessions"))
	}

	return c.JSON(dto.Success(sessions))
}

// DeleteSession 删除指定会话
// @Summary 删除会话
// @Description 删除指定的设备会话
// @Tags 认证
// @Security BearerAuth
// @Param sessionId path string true "会话ID"
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/sessions/{sessionId} [delete]
func (h *AuthHandler) DeleteSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID := c.Params("sessionId")

	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "session ID is required"))
	}

	if err := h.authService.DeleteSession(c.Context(), userID, sessionID); err != nil {
		h.logger.Error("failed to delete session",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.String("session_id", sessionID),
		)

		statusCode := fiber.StatusInternalServerError
		message := "failed to delete session"

		errMsg := err.Error()
		switch errMsg {
		case "session not found":
			statusCode = fiber.StatusNotFound
			message = errMsg
		case "session does not belong to user":
			statusCode = fiber.StatusForbidden
			message = "access denied"
		}

		return c.Status(statusCode).JSON(dto.Error(statusCode, message))
	}

	return c.JSON(dto.Success("session deleted successfully"))
}

// GetActiveSessions 获取活跃会话
// @Summary 获取活跃会话
// @Description 获取用户所有活跃的会话
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.SessionListResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/sessions/active [get]
func (h *AuthHandler) GetActiveSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	sessions, err := h.authService.GetActiveSessions(c.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get active sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get active sessions"))
	}

	return c.JSON(dto.Success(sessions))
}

// GetRecentSessions 获取最近会话
// @Summary 获取最近会话
// @Description 获取用户最近指定天数内的会话
// @Tags 认证
// @Security BearerAuth
// @Param days query int false "天数，默认7天"
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.SessionListResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/sessions/recent [get]
func (h *AuthHandler) GetRecentSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取天数参数，默认7天
	days := c.QueryInt("days", 7)
	if days <= 0 || days > 365 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "days must be between 1 and 365"))
	}

	sessions, err := h.authService.GetRecentSessions(c.Context(), userID, days)
	if err != nil {
		h.logger.Error("failed to get recent sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
			interfaces.Int("days", days),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get recent sessions"))
	}

	return c.JSON(dto.Success(sessions))
}

// GetSessionStats 获取会话统计
// @Summary 获取会话统计
// @Description 获取用户会话的统计信息
// @Tags 认证
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APIResponse{data=map[string]int64}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/sessions/stats [get]
func (h *AuthHandler) GetSessionStats(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// 获取总会话数
	totalSessions, err := h.authService.CountUserSessions(c.Context(), userID, false)
	if err != nil {
		h.logger.Error("failed to count total sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get session stats"))
	}

	// 获取活跃会话数
	activeSessions, err := h.authService.CountUserSessions(c.Context(), userID, true)
	if err != nil {
		h.logger.Error("failed to count active sessions",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "failed to get session stats"))
	}

	stats := map[string]int64{
		"total_sessions":    totalSessions,
		"active_sessions":   activeSessions,
		"inactive_sessions": totalSessions - activeSessions,
	}

	return c.JSON(dto.Success(stats))
}

// ============================================================================
// Token黑名单管理API
// ============================================================================

// BlacklistToken 将token添加到黑名单
// @Summary 将token添加到黑名单
// @Description 撤销指定的访问token，使其失效
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.BlacklistTokenRequest true "Token黑名单请求"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/blacklist-token [post]
// @Security Bearer
func (h *AuthHandler) BlacklistToken(c *fiber.Ctx) error {
	var req dto.BlacklistTokenRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("failed to parse blacklist token request",
			interfaces.ErrorField(err),
			interfaces.String("request_id", middleware.GetRequestID(c)),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "Invalid request format"))
	}

	// 验证请求参数
	if err := h.validator.ValidateStruct(&req); err != nil {
		h.logger.Warn("blacklist token validation failed",
			interfaces.ErrorField(err),
			interfaces.String("request_id", middleware.GetRequestID(c)),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.Error(400, "Validation failed: "+err.Error()))
	}

	// 将token添加到黑名单
	if err := h.authService.BlacklistToken(c.Context(), req.Token); err != nil {
		h.logger.Error("failed to blacklist token",
			interfaces.ErrorField(err),
			interfaces.String("request_id", middleware.GetRequestID(c)),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to blacklist token"))
	}

	h.logger.Info("token blacklisted successfully",
		interfaces.String("reason", req.Reason),
		interfaces.String("request_id", middleware.GetRequestID(c)),
	)

	return c.JSON(dto.Success("Token has been successfully blacklisted"))
}

// GetBlacklistStats 获取黑名单统计信息
// @Summary 获取黑名单统计信息
// @Description 获取token黑名单的统计信息
// @Tags 认证
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.BlacklistStatsResponse}
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/blacklist-stats [get]
// @Security Bearer
func (h *AuthHandler) GetBlacklistStats(c *fiber.Ctx) error {
	stats, err := h.authService.GetBlacklistStats(c.Context())
	if err != nil {
		h.logger.Error("failed to get blacklist stats",
			interfaces.ErrorField(err),
			interfaces.String("request_id", middleware.GetRequestID(c)),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to get blacklist stats"))
	}

	// 转换为响应DTO
	response := dto.BlacklistStatsResponse{
		TotalBlacklistedTokens: stats["total_blacklisted_tokens"].(int64),
		LastUpdated:            stats["last_updated"].(time.Time),
	}

	return c.JSON(dto.Success(response))
}

// ClearExpiredBlacklistTokens 清理过期的黑名单token
// @Summary 清理过期的黑名单token
// @Description 手动清理过期的黑名单token
// @Tags 认证
// @Produce json
// @Success 200 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/blacklist-cleanup [post]
// @Security Bearer
func (h *AuthHandler) ClearExpiredBlacklistTokens(c *fiber.Ctx) error {
	if err := h.authService.ClearExpiredBlacklistTokens(c.Context()); err != nil {
		h.logger.Error("failed to clear expired blacklist tokens",
			interfaces.ErrorField(err),
			interfaces.String("request_id", middleware.GetRequestID(c)),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.Error(500, "Failed to clear expired tokens"))
	}

	h.logger.Info("expired blacklist tokens cleared successfully",
		interfaces.String("request_id", middleware.GetRequestID(c)),
	)

	return c.JSON(dto.Success("Expired blacklist tokens cleared successfully"))
}
