package rbac

import (
	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	svcIf "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: PermissionsQueryHandler
//
// Description: 有效权限查询处理器。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionsQueryHandler struct {
	svc    svcIf.PermissionsQueryService
	logger interfaces.Logger
}

// Function: NewPermissionsQueryHandler
//
// Description: 创建并返回 PermissionsQueryHandler 实例。
//
// Parameters:
//
//	svc(svcIf.PermissionsQueryService): 权限查询服务
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	(*PermissionsQueryHandler): 处理器实例
func NewPermissionsQueryHandler(svc svcIf.PermissionsQueryService, logger interfaces.Logger) *PermissionsQueryHandler {
	return &PermissionsQueryHandler{svc: svc, logger: logger}
}

// Function: GetEffectivePermissions
//
// Description: 查询并返回给定用户的有效权限列表。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 处理错误
func (h *PermissionsQueryHandler) GetEffectivePermissions(c *fiber.Ctx) error {
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("userId is required"))
	}
	perms, err := h.svc.GetEffectivePermissions(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError(err.Error()))
	}
	return c.JSON(dto.Success(perms))
}
