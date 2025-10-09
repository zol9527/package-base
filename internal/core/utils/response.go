package utils

import "github.com/gofiber/fiber/v2"

// Function: Success
//
// Description: 返回统一成功响应的 JSON 封装，包含 code=0、message 及数据对象。
//
// Parameters:
//
//	c(*fiber.Ctx): 请求上下文
//	data(interface{}): 需要返回的业务数据，可为 nil
//	message(string): 可选的提示信息，空则默认为 ok
//
// Returns:
//
//	error(error): fiber JSON 写入产生的错误（若有）
func Success(c *fiber.Ctx, data interface{}, message string) error {
	if message == "" {
		message = "ok"
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": message,
		"data":    data,
	})
}

// Function: Error
//
// Description: 返回统一失败响应的 JSON 封装，包含非 0 code、错误 message 与可选 details。
//
// Parameters:
//
//	c(*fiber.Ctx): 请求上下文
//	http_status(int): HTTP 状态码
//	code(string): 业务错误码，采用 AUTH_xxx 等前缀规范
//	message(string): 人类可读错误描述
//	details(interface{}): 额外调试信息（生产可置 nil）
//
// Returns:
//
//	error(error): fiber JSON 写入错误（若有）
func Error(c *fiber.Ctx, http_status int, code string, message string, details interface{}) error {
	if http_status == 0 {
		http_status = fiber.StatusBadRequest
	}
	if code == "" {
		code = "AUTH_UNKNOWN_ERROR"
	}
	return c.Status(http_status).JSON(fiber.Map{
		"code":    code,
		"message": message,
		"details": details,
	})
}
