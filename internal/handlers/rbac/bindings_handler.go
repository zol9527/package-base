package rbac

import (
	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	svcIf "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: BindingsHandler
//
// Description: 用户-角色绑定相关处理器。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type BindingsHandler struct {
	bindingService svcIf.BindingService
	logger         interfaces.Logger
}

// Function: NewBindingsHandler
//
// Description: 创建并返回 BindingsHandler 实例。
//
// Parameters:
//
//	bindingService(svcIf.BindingService): 绑定服务
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	(*BindingsHandler): 处理器实例
func NewBindingsHandler(bindingService svcIf.BindingService, logger interfaces.Logger) *BindingsHandler {
	return &BindingsHandler{bindingService: bindingService, logger: logger}
}

// Function: UpdateUserRoles
//
// Description: 覆盖绑定用户在指定 scope 下的角色集合。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 处理错误
func (h *BindingsHandler) UpdateUserRoles(c *fiber.Ctx) error {
	userID := c.Params("userId")
	var req dto.UserBindRolesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid request body"))
	}
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("userId is required"))
	}
	if len(req.RoleIDs) == 0 {
		// 允许解绑所有角色，但需显式传空数组；如果 body 为空则算 bad request
		if req.ExplicitEmpty != nil && *req.ExplicitEmpty {
			// 显式空：解除绑定
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("role_ids is required (use [] to unbind)"))
		}
	}
	if err := h.bindingService.BindUserRoles(c.Context(), userID, req.RoleIDs, req.Scope); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to bind roles"))
	}
	return c.JSON(dto.Success("roles bound"))
}
