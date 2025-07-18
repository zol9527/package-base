package handlers

import (
	"github.com/gofiber/fiber/v2"

	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	serviceInterfaces "package-base/internal/services/interfaces"
)

// TotpHandler 2FA TOTP处理器
type TotpHandler struct {
	totpService serviceInterfaces.TotpService
	logger      interfaces.Logger
}

// NewTotpHandler 创建TOTP处理器
func NewTotpHandler(totpService serviceInterfaces.TotpService, logger interfaces.Logger) *TotpHandler {
	return &TotpHandler{
		totpService: totpService,
		logger:      logger,
	}
}

// GenerateSecret 生成2FA设置信息
// @Summary 生成2FA设置信息
// @Description 为用户生成TOTP密钥和QR码
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.TotpSetupResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/setup [post]
func (h *TotpHandler) GenerateSecret(c *fiber.Ctx) error {
	// 从中间件获取用户ID（这里需要认证中间件支持）
	userID := c.Locals("user_id").(string)

	// 获取用户信息作为账户名（这里可以从用户服务获取）
	accountName := "user@example.com" // 实际应该从用户信息获取

	setupInfo, err := h.totpService.GenerateSecret(c.Context(), userID, accountName)
	if err != nil {
		h.logger.Error("生成2FA设置信息失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"生成2FA设置信息失败",
		))
	}

	response := dto.TotpSetupResponse{
		Secret:      setupInfo.Secret,
		QRCodeB64:   setupInfo.QRCodeB64,
		URL:         setupInfo.URL,
		Issuer:      "AuthSystem",
		AccountName: accountName,
	}

	h.logger.Info("成功生成2FA设置信息",
		interfaces.String("user_id", userID),
	)

	return c.JSON(dto.SuccessWithMessage("成功生成2FA设置信息", response))
}

// EnableTOTP 启用2FA
// @Summary 启用2FA
// @Description 验证TOTP令牌并启用2FA
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.EnableTOTPRequest true "启用2FA请求"
// @Success 200 {object} dto.APIResponse{data=dto.EnableTOTPResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/enable [post]
func (h *TotpHandler) EnableTOTP(c *fiber.Ctx) error {
	var req dto.EnableTOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"请求参数解析失败",
		))
	}

	userID := c.Locals("user_id").(string)

	totpSecret, backupCodes, err := h.totpService.EnableTOTP(c.Context(), userID, req.Secret, req.Token)
	if err != nil {
		h.logger.Error("启用2FA失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"启用2FA失败",
		))
	}

	response := dto.EnableTOTPResponse{
		Success:     true,
		BackupCodes: backupCodes,
		Message:     "2FA已成功启用，请妥善保存备用码",
	}

	h.logger.Info("成功启用2FA",
		interfaces.String("user_id", userID),
		interfaces.String("totp_id", totpSecret.ID),
	)

	return c.JSON(dto.SuccessWithMessage("2FA启用成功", response))
}

// DisableTOTP 禁用2FA
// @Summary 禁用2FA
// @Description 验证TOTP令牌并禁用2FA
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.DisableTOTPRequest true "禁用2FA请求"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/disable [post]
func (h *TotpHandler) DisableTOTP(c *fiber.Ctx) error {
	var req dto.DisableTOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"请求参数解析失败",
		))
	}

	userID := c.Locals("user_id").(string)

	if err := h.totpService.DisableTOTP(c.Context(), userID, req.Token); err != nil {
		h.logger.Error("禁用2FA失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"禁用2FA失败",
		))
	}

	h.logger.Info("成功禁用2FA",
		interfaces.String("user_id", userID),
	)

	return c.JSON(dto.SuccessWithMessage("2FA禁用成功", nil))
}

// VerifyTOTP 验证TOTP令牌
// @Summary 验证TOTP令牌
// @Description 验证用户提供的TOTP令牌
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.VerifyTOTPRequest true "验证TOTP请求"
// @Success 200 {object} dto.APIResponse{data=dto.VerifyResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/verify [post]
func (h *TotpHandler) VerifyTOTP(c *fiber.Ctx) error {
	var req dto.VerifyTOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"请求参数解析失败",
		))
	}

	userID := c.Locals("user_id").(string)

	valid, err := h.totpService.VerifyTOTP(c.Context(), userID, req.Token)
	if err != nil {
		h.logger.Error("验证TOTP令牌失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"验证TOTP令牌失败",
		))
	}

	response := dto.VerifyResponse{
		Valid:   valid,
		Message: map[bool]string{true: "TOTP令牌验证成功", false: "TOTP令牌无效"}[valid],
	}

	h.logger.Info("TOTP令牌验证结果",
		interfaces.String("user_id", userID),
		interfaces.Bool("valid", valid),
	)

	return c.JSON(dto.SuccessWithMessage("验证完成", response))
}

// VerifyBackupCode 验证备用码
// @Summary 验证备用码
// @Description 验证用户提供的备用码
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.VerifyBackupCodeRequest true "验证备用码请求"
// @Success 200 {object} dto.APIResponse{data=dto.VerifyResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/backup-code/verify [post]
func (h *TotpHandler) VerifyBackupCode(c *fiber.Ctx) error {
	var req dto.VerifyBackupCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(
			"请求参数解析失败",
		))
	}

	userID := c.Locals("user_id").(string)

	valid, err := h.totpService.VerifyBackupCode(c.Context(), userID, req.Code)
	if err != nil {
		h.logger.Error("验证备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"验证备用码失败",
		))
	}

	response := dto.VerifyResponse{
		Valid:   valid,
		Message: map[bool]string{true: "备用码验证成功", false: "备用码无效或已使用"}[valid],
	}

	h.logger.Info("备用码验证结果",
		interfaces.String("user_id", userID),
		interfaces.Bool("valid", valid),
	)

	return c.JSON(dto.SuccessWithMessage("验证完成", response))
}

// RegenerateBackupCodes 重新生成备用码
// @Summary 重新生成备用码
// @Description 为用户重新生成备用码
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.RegenerateBackupCodesResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/backup-codes/regenerate [post]
func (h *TotpHandler) RegenerateBackupCodes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	backupCodes, err := h.totpService.RegenerateBackupCodes(c.Context(), userID)
	if err != nil {
		h.logger.Error("重新生成备用码失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"重新生成备用码失败",
		))
	}

	response := dto.RegenerateBackupCodesResponse{
		BackupCodes: backupCodes,
		Count:       len(backupCodes),
		Message:     "备用码已重新生成，请妥善保存",
	}

	h.logger.Info("成功重新生成备用码",
		interfaces.String("user_id", userID),
		interfaces.Int("count", len(backupCodes)),
	)

	return c.JSON(dto.SuccessWithMessage("备用码重新生成成功", response))
}

// GetTOTPStatus 获取2FA状态
// @Summary 获取2FA状态
// @Description 获取用户的2FA启用状态
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.TotpStatusResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/status [get]
func (h *TotpHandler) GetTOTPStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	status, err := h.totpService.GetUserTOTPStatus(c.Context(), userID)
	if err != nil {
		h.logger.Error("获取2FA状态失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"获取2FA状态失败",
		))
	}

	response := dto.TotpStatusResponse{
		IsEnabled: status.IsEnabled,
		Issuer:    status.Issuer,
		CreatedAt: status.CreatedAt,
	}

	return c.JSON(dto.SuccessWithMessage("获取2FA状态成功", response))
}

// GetBackupCodeStatus 获取备用码状态
// @Summary 获取备用码状态
// @Description 获取用户备用码的使用状态
// @Tags 2FA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.APIResponse{data=dto.BackupCodeStatusResponse}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /api/auth/2fa/backup-codes/status [get]
func (h *TotpHandler) GetBackupCodeStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	status, err := h.totpService.GetBackupCodeStatus(c.Context(), userID)
	if err != nil {
		h.logger.Error("获取备用码状态失败",
			interfaces.ErrorField(err),
			interfaces.String("user_id", userID),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(
			"获取备用码状态失败",
		))
	}

	response := dto.BackupCodeStatusResponse{
		TotalCount:  status.TotalCount,
		UsedCount:   status.UsedCount,
		UnusedCount: status.UnusedCount,
	}

	return c.JSON(dto.SuccessWithMessage("获取备用码状态成功", response))
}
