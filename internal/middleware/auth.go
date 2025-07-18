package middleware

import (
	"strings"

	"package-base/internal/core/interfaces"
	"package-base/internal/core/utils"
	"package-base/internal/dto"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware JWT认证中间件配置
type AuthMiddleware struct {
	jwtManager *utils.JWTManager
	logger     interfaces.Logger
}

// NewAuthMiddleware 创建JWT认证中间件
func NewAuthMiddleware(jwtSecret string, logger interfaces.Logger) *AuthMiddleware {
	// 这里只是为了验证Token，所以duration可以设置为任意值
	jwtManager := utils.NewJWTManager(jwtSecret, 0, 0)

	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Authenticate JWT认证中间件函数
func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. 从Header中获取Authorization
		auth := c.Get("Authorization")
		if auth == "" {
			m.logger.Warn("missing authorization header",
				interfaces.String("path", c.Path()),
				interfaces.String("method", c.Method()),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Error(401, "missing authorization header"))
		}

		// 2. 检查Bearer格式
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(auth, bearerPrefix) {
			m.logger.Warn("invalid authorization format",
				interfaces.String("auth", auth),
				interfaces.String("path", c.Path()),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Error(401, "invalid authorization format"))
		}

		// 3. 提取Token
		token := strings.TrimPrefix(auth, bearerPrefix)
		if token == "" {
			m.logger.Warn("empty token",
				interfaces.String("path", c.Path()),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Error(401, "empty token"))
		}

		// 4. 验证Token
		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			m.logger.Warn("invalid token",
				interfaces.ErrorField(err),
				interfaces.String("path", c.Path()),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(dto.Error(401, "invalid token"))
		}

		// 5. 将用户信息存储到Context中
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		// 6. 记录认证成功日志
		m.logger.Debug("authentication successful",
			interfaces.String("user_id", claims.UserID),
			interfaces.String("username", claims.Username),
			interfaces.String("path", c.Path()),
			interfaces.String("method", c.Method()),
		)

		return c.Next()
	}
}

// Optional 可选认证中间件 - 如果有Token则验证，没有Token则跳过
func (m *AuthMiddleware) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. 从Header中获取Authorization
		auth := c.Get("Authorization")
		if auth == "" {
			// 没有认证头，继续处理
			return c.Next()
		}

		// 2. 检查Bearer格式
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(auth, bearerPrefix) {
			// 格式错误，继续处理但不设置用户信息
			return c.Next()
		}

		// 3. 提取Token
		token := strings.TrimPrefix(auth, bearerPrefix)
		if token == "" {
			// Token为空，继续处理
			return c.Next()
		}

		// 4. 验证Token
		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			// Token无效，继续处理但不设置用户信息
			m.logger.Debug("optional auth: invalid token",
				interfaces.ErrorField(err),
				interfaces.String("path", c.Path()),
			)
			return c.Next()
		}

		// 5. 将用户信息存储到Context中
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		m.logger.Debug("optional authentication successful",
			interfaces.String("user_id", claims.UserID),
			interfaces.String("path", c.Path()),
		)

		return c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *AuthMiddleware) RequireRole(requiredRole int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取用户角色
		userRole, ok := c.Locals("role").(int)
		if !ok {
			m.logger.Warn("role not found in context",
				interfaces.String("path", c.Path()),
			)
			return c.Status(fiber.StatusForbidden).JSON(dto.Error(403, "access denied"))
		}

		// 检查角色权限
		if userRole < requiredRole {
			userID := c.Locals("user_id").(string)
			m.logger.Warn("insufficient role permissions",
				interfaces.String("user_id", userID),
				interfaces.Int("user_role", userRole),
				interfaces.Int("required_role", requiredRole),
				interfaces.String("path", c.Path()),
			)
			return c.Status(fiber.StatusForbidden).JSON(dto.Error(403, "insufficient permissions"))
		}

		return c.Next()
	}
}

// GetUserID 从Context中获取用户ID的辅助函数
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUsername 从Context中获取用户名的辅助函数
func GetUsername(c *fiber.Ctx) string {
	username, ok := c.Locals("username").(string)
	if !ok {
		return ""
	}
	return username
}

// GetUserEmail 从Context中获取用户邮箱的辅助函数
func GetUserEmail(c *fiber.Ctx) string {
	email, ok := c.Locals("email").(string)
	if !ok {
		return ""
	}
	return email
}

// GetUserRole 从Context中获取用户角色的辅助函数
func GetUserRole(c *fiber.Ctx) int {
	role, ok := c.Locals("role").(int)
	if !ok {
		return 0
	}
	return role
}

// IsAuthenticated 检查用户是否已认证的辅助函数
func IsAuthenticated(c *fiber.Ctx) bool {
	return GetUserID(c) != ""
}
