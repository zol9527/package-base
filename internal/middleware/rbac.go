package middleware

import (
	coreInterfaces "package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: RBAC
//
// Description: RBAC 鉴权中间件骨架：读取用户身份、调用服务做权限判定、对未授权返回 401/403 并记录审计。
//
// Parameters:
//
//    service interfaces.AuthorizationService: 鉴权服务接口（后续定义于 internal/services/interfaces）
//    logger  interfaces.Logger: 统一日志接口
//
// Returns:
//
//    fiber.Handler: Fiber 中间件处理函数
func RBAC(service serviceInterfaces.AuthorizationService, logger coreInterfaces.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 占位：从上下文读取用户身份（如 JWT 解析结果），若无则 401
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return utils.Error(c, fiber.StatusUnauthorized, utils.RBAC_UNAUTHORIZED, "unauthorized", nil)
		}

		// 占位：资源与动作从路由元数据或约定中获取，这里先占位
	resource := c.Path()
	action := c.Method()

	allowed, reason := false, "not implemented"
	_ = resource
	_ = action
	// TODO: 调用 service.Authorize(c.Context(), userID.(string), resource, action)

		if !allowed {
			// 记录审计可由 service 负责，这里仅返回统一错误
			return utils.Error(c, fiber.StatusForbidden, utils.RBAC_FORBIDDEN, "forbidden", map[string]string{"reason": reason})
		}
		return c.Next()
	}
}
