package middleware

import (
	"time"

	"package-base/internal/core/interfaces"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct {
	logger interfaces.Logger
}

// NewLogger 创建日志中间件
func NewLogger(logger interfaces.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// Handler 日志中间件处理函数
func (m *LoggerMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 开始时间
		start := time.Now()

		// 生成请求ID
		requestID := uuid.New().String()
		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		// 记录请求开始日志
		m.logger.Info("request started",
			interfaces.String("request_id", requestID),
			interfaces.String("method", c.Method()),
			interfaces.String("path", c.Path()),
			interfaces.String("ip", c.IP()),
			interfaces.String("user_agent", c.Get("User-Agent")),
		)

		// 执行下一个处理器
		err := c.Next()

		// 计算处理时间
		duration := time.Since(start)
		c.Set("X-Response-Time", duration.String())

		// 获取用户信息（如果有）
		userID := GetUserID(c)
		username := GetUsername(c)

		// 记录请求完成日志
		logFields := []interfaces.Field{
			interfaces.String("request_id", requestID),
			interfaces.String("method", c.Method()),
			interfaces.String("path", c.Path()),
			interfaces.String("ip", c.IP()),
			interfaces.Int("status", c.Response().StatusCode()),
			interfaces.String("duration", duration.String()),
			interfaces.String("user_agent", c.Get("User-Agent")),
		}

		// 添加用户信息（如果存在）
		if userID != "" {
			logFields = append(logFields, interfaces.String("user_id", userID))
		}
		if username != "" {
			logFields = append(logFields, interfaces.String("username", username))
		}

		// 根据状态码决定日志级别
		statusCode := c.Response().StatusCode()
		if err != nil {
			logFields = append(logFields, interfaces.ErrorField(err))
			m.logger.Error("request completed with error", logFields...)
		} else if statusCode >= 500 {
			m.logger.Error("request completed with server error", logFields...)
		} else if statusCode >= 400 {
			m.logger.Warn("request completed with client error", logFields...)
		} else {
			m.logger.Info("request completed successfully", logFields...)
		}

		return err
	}
}

// Recovery 错误恢复中间件
func (m *LoggerMiddleware) Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID := GetRequestID(c)
				userID := GetUserID(c)

				m.logger.Error("panic recovered",
					interfaces.String("request_id", requestID),
					interfaces.String("user_id", userID),
					interfaces.String("method", c.Method()),
					interfaces.String("path", c.Path()),
					interfaces.String("ip", c.IP()),
					interfaces.Any("panic", r),
				)

				// 返回500错误
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    500,
					"message": "internal server error",
				})
			}
		}()

		return c.Next()
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 设置安全头
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		return c.Next()
	}
}

// GetRequestID 从Context中获取请求ID的辅助函数
func GetRequestID(c *fiber.Ctx) string {
	requestID, ok := c.Locals("request_id").(string)
	if !ok {
		return ""
	}
	return requestID
}
