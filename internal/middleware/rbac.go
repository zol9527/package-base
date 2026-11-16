package middleware

import (
	"context"
	coreInterfaces "package-base/internal/core/interfaces"
	"package-base/internal/dto"
	serviceInterfaces "package-base/internal/services/interfaces"

	"github.com/gofiber/fiber/v2"
)

// Function: RBAC
//
// Description: RBAC 鉴权中间件骨架：读取用户身份、调用服务做权限判定、对未授权返回 401/403 并记录审计。
//
// Parameters:
//
//	service interfaces.AuthorizationService: 鉴权服务接口（后续定义于 internal/services/interfaces）
//	logger  interfaces.Logger: 统一日志接口
//
// Returns:
//
//	fiber.Handler: Fiber 中间件处理函数
func RBAC(service serviceInterfaces.AuthorizationService, logger coreInterfaces.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 读取用户身份
		v := c.Locals("user_id")
		userID, _ := v.(string)
		if userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Unauthorized("unauthorized"))
		}

		// 提取资源与动作（后续可替换为更语义化的资源名）
		resource := c.Path()
		action := c.Method()

		allowed, reason := service.Authorize(c.Context(), userID, resource, action)
		if !allowed {
			// 审计（可忽略错误）
			_ = tryAudit(service, c, userID, resource, action, false, reason)
			return c.Status(fiber.StatusForbidden).JSON(dto.Forbidden("forbidden"))
		}
		// 允许时可按需记录审计
		_ = tryAudit(service, c, userID, resource, action, true, "")
		return c.Next()
	}
}

// Function: tryAudit
//
// Description: 尝试调用服务审计记录，忽略审计错误。
//
// Parameters:
//
//	service(serviceInterfaces.AuthorizationService): 鉴权服务
//	c(*fiber.Ctx): 上下文
//	userID(string): 用户ID
//	resource(string): 资源
//	action(string): 动作
//	allowed(bool): 是否允许
//	reason(string): 原因
//
// Returns:
//
//	error(error): 审计错误（已忽略）
func tryAudit(service serviceInterfaces.AuthorizationService, c *fiber.Ctx, userID, resource, action string, allowed bool, reason string) error {
	// 复用 request id 作为 requestID（优先从 Fiber requestid 中间件的 Locals 提取）
	var reqID string
	if v := c.Locals("requestid"); v != nil {
		if s, ok := v.(string); ok {
			reqID = s
		}
	}
	if reqID == "" {
		// 回退从请求头读取（若客户端自带）
		reqID = c.Get("X-Request-ID")
	}
	client := c.IP()
	// 如果实现了 AuditDecision 方法则记录；接口中未包含该方法，这里通过类型断言可选支持
	type auditCap interface {
		AuditDecision(ctx context.Context, userID, resource, action string, allowed bool, reason, requestID, client string) error
	}
	if svc, ok := any(service).(auditCap); ok {
		_ = svc.AuditDecision(c.Context(), userID, resource, action, allowed, reason, reqID, client)
	}
	return nil
}
