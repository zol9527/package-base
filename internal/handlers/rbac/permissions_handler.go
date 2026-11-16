package rbac

import (
	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	model "package-base/internal/models/rbac"
	repoIf "package-base/internal/repositories/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: PermissionsHandler
//
// Description: 权限管理处理器（当前仅实现创建）。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type PermissionsHandler struct {
	repo   repoIf.PermissionRepository
	logger interfaces.Logger
}

// Function: NewPermissionsHandler
//
// Description: 创建并返回 PermissionsHandler。
//
// Parameters:
//
//	repo(repoIf.PermissionRepository): 权限仓储
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	(*PermissionsHandler): 处理器实例
func NewPermissionsHandler(repo repoIf.PermissionRepository, logger interfaces.Logger) *PermissionsHandler {
	return &PermissionsHandler{repo: repo, logger: logger}
}

// Function: CreatePermission
//
// Description: 创建权限实体。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 错误
func (h *PermissionsHandler) CreatePermission(c *fiber.Ctx) error {
	var req struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
		Effect   string `json:"effect"`
		Note     string `json:"note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid request body"))
	}
	if req.Resource == "" || req.Action == "" || req.Effect == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("resource/action/effect are required"))
	}
	perm := &model.Permission{Resource: req.Resource, Action: req.Action, Effect: req.Effect, Note: req.Note}
	if err := h.repo.Create(c.Context(), perm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to create permission"))
	}
	return c.JSON(dto.Success(fiber.Map{"id": perm.ID}))
}
