package rbac

import (
	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	rbacModel "package-base/internal/models/rbac"
	svcIf "package-base/internal/services/interfaces"

	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Function: RolesHandler
//
// Description: 角色管理相关的 HTTP 处理器，提供创建等接口。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type RolesHandler struct {
	roleService svcIf.RoleService
	logger      interfaces.Logger
}

// Function: NewRolesHandler
//
// Description: 创建并返回 RolesHandler 实例。
//
// Parameters:
//
//	roleService(svcIf.RoleService): 角色服务
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	(*RolesHandler): 处理器实例
func NewRolesHandler(roleService svcIf.RoleService, logger interfaces.Logger) *RolesHandler {
	return &RolesHandler{roleService: roleService, logger: logger}
}

// Function: CreateRole
//
// Description: 创建角色 API。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 处理错误
func (h *RolesHandler) CreateRole(c *fiber.Ctx) error {
	var req dto.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid request body"))
	}
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("name is required"))
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	m := &rbacModel.Role{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     enabled,
	}
	created, err := h.roleService.CreateRole(c.Context(), m)
	if err != nil {
		switch err.Error() {
		case "role name already exists":
			return c.Status(fiber.StatusConflict).JSON(dto.Error(409, err.Error()))
		case "role name is required":
			return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest(err.Error()))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to create role"))
		}
	}

	resp := dto.RoleResponse{
		ID:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		BuiltIn:     created.BuiltIn,
		Enabled:     created.Enabled,
	}
	return c.JSON(dto.SuccessWithMessage("role created", resp))
}

// Function: UpdateRole
//
// Description: 更新角色信息。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 处理错误
func (h *RolesHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid id"))
	}
	var req dto.RoleCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid request body"))
	}
	m := &rbacModel.Role{Name: req.Name, Description: req.Description}
	if req.Enabled != nil {
		m.Enabled = *req.Enabled
	}
	updated, err := h.roleService.UpdateRole(c.Context(), id, m)
	if err != nil {
		switch err.Error() {
		case "role not found":
			return c.Status(fiber.StatusNotFound).JSON(dto.NotFound(err.Error()))
		case "built-in role cannot be modified":
			return c.Status(fiber.StatusForbidden).JSON(dto.Forbidden(err.Error()))
		case "role name already exists":
			return c.Status(fiber.StatusConflict).JSON(dto.Error(409, err.Error()))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to update role"))
		}
	}
	resp := dto.RoleResponse{ID: updated.ID, Name: updated.Name, Description: updated.Description, BuiltIn: updated.BuiltIn, Enabled: updated.Enabled}
	return c.JSON(dto.SuccessWithMessage("role updated", resp))
}

// Function: DeleteRole
//
// Description: 删除角色。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//
// Returns:
//
//	error: 处理错误
func (h *RolesHandler) DeleteRole(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.BadRequest("invalid id"))
	}
	if err := h.roleService.DeleteRole(c.Context(), id); err != nil {
		switch err.Error() {
		case "role not found":
			return c.Status(fiber.StatusNotFound).JSON(dto.NotFound(err.Error()))
		case "built-in role cannot be deleted":
			return c.Status(fiber.StatusForbidden).JSON(dto.Forbidden(err.Error()))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to delete role"))
		}
	}
	return c.JSON(dto.Success("role deleted"))
}

// Function: parseUintParam
//
// Description: 解析路径中的无符号整数参数。
//
// Parameters:
//
//	c(*fiber.Ctx): Fiber 上下文
//	name(string): 参数名
//
// Returns:
//
//	(uint, error): 解析后的数值与可能的错误
func parseUintParam(c *fiber.Ctx, name string) (uint, error) {
	v := c.Params(name)
	if v == "" {
		return 0, fiber.ErrBadRequest
	}
	var id uint64
	var err error
	if id, err = strconv.ParseUint(v, 10, 64); err != nil {
		return 0, err
	}
	return uint(id), nil
}
