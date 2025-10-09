package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Function: TraceMiddleware
//
// Description: 生成并注入 trace_id (与 request_id 区分) 到请求上下文与响应头，供日志与客户端关联调用链。
//
// Parameters:
//
//	无（结构体定义）
//
// Returns:
//
//	无（结构体定义）
type TraceMiddleware struct{}

// Function: NewTraceMiddleware
//
// Description: 创建 TraceMiddleware 实例。
//
// Parameters:
//
//	无
//
// Returns:
//
//	m(*TraceMiddleware): 中间件实例
func NewTraceMiddleware() *TraceMiddleware { return &TraceMiddleware{} }

// Function: Handler
//
// Description: Fiber 处理函数，若请求头已携带 X-Trace-ID 则复用，否则生成新的 UUID。
//
// Parameters:
//
//	无（方法使用 *fiber.Ctx）
//
// Returns:
//
//	fiber.Handler: 处理函数
func (m *TraceMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Locals("trace_id", traceID)
		c.Set("X-Trace-ID", traceID)
		return c.Next()
	}
}

// Function: GetTraceID
//
// Description: 从上下文获取 trace_id。
//
// Parameters:
//
//	c(*fiber.Ctx): 请求上下文
//
// Returns:
//
//	traceID(string): trace_id 字符串
func GetTraceID(c *fiber.Ctx) string {
	v, ok := c.Locals("trace_id").(string)
	if !ok {
		return ""
	}
	return v
}
