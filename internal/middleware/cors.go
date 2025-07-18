package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig 默认CORS配置
var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{
		fiber.MethodGet,
		fiber.MethodPost,
		fiber.MethodPut,
		fiber.MethodDelete,
		fiber.MethodOptions,
		fiber.MethodPatch,
	},
	AllowHeaders: []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-Requested-With",
		"X-Request-ID",
	},
	AllowCredentials: true,
	ExposeHeaders: []string{
		"X-Request-ID",
		"X-Response-Time",
	},
	MaxAge: 86400, // 24小时
}

// NewCORS 创建CORS中间件
func NewCORS(config ...CORSConfig) fiber.Handler {
	// 使用提供的配置或默认配置
	cfg := DefaultCORSConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return cors.New(cors.Config{
		AllowOrigins:     joinStrings(cfg.AllowOrigins, ","),
		AllowMethods:     joinStrings(cfg.AllowMethods, ","),
		AllowHeaders:     joinStrings(cfg.AllowHeaders, ","),
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    joinStrings(cfg.ExposeHeaders, ","),
		MaxAge:           cfg.MaxAge,
	})
}

// NewCORSForDevelopment 开发环境CORS配置
func NewCORSForDevelopment() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Request-ID",
		AllowCredentials: true,
		ExposeHeaders:    "X-Request-ID,X-Response-Time",
		MaxAge:           86400,
	})
}

// NewCORSForProduction 生产环境CORS配置
func NewCORSForProduction(allowedOrigins []string) fiber.Handler {
	origins := "*"
	if len(allowedOrigins) > 0 {
		origins = joinStrings(allowedOrigins, ",")
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           3600, // 1小时
	})
}

// joinStrings 连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
