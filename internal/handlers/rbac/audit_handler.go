package rbac

import (
	"strconv"

	"package-base/internal/core/interfaces"
	"package-base/internal/dto"
	repoIf "package-base/internal/repositories/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: AuditHandler
//
// Description: 审计记录查询处理器。
//
// Parameters:
//
//	无
//
// Returns:
//
//	无
type AuditHandler struct {
	repo   repoIf.AuditRepository
	logger interfaces.Logger
}

// Function: NewAuditHandler
//
// Description: 创建审计查询处理器。
//
// Parameters:
//
//	repo(repoIf.AuditRepository): 审计仓储
//	logger(interfaces.Logger): 日志
//
// Returns:
//
//	(*AuditHandler): 处理器实例
func NewAuditHandler(repo repoIf.AuditRepository, logger interfaces.Logger) *AuditHandler {
	return &AuditHandler{repo: repo, logger: logger}
}

// Function: Query
//
// Description: 查询审计记录，支持 userId/resource/action/limit 过滤。
//
// Parameters:
//
//	c(*fiber.Ctx): 上下文
//
// Returns:
//
//	error: 错误
func (h *AuditHandler) Query(c *fiber.Ctx) error {
	userID := c.Query("userId")
	resource := c.Query("resource")
	action := c.Query("action")
	limitStr := c.Query("limit")
	limit := 100
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	rows, err := h.repo.QueryDecisions(c.Context(), userID, resource, action, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.InternalServerError("failed to query audit"))
	}
	return c.JSON(dto.Success(rows))
}
