package dto

// ============================================================================
// 通用响应结构
// ============================================================================

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Pages    int         `json:"pages"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users    []UserInfo `json:"users"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Pages    int        `json:"pages"`
}

// ============================================================================
// 响应构建器
// ============================================================================

// Success 成功响应
// Success 成功响应 (统一业务成功码=0)
func Success(data interface{}) *APIResponse {
	return &APIResponse{Code: 0, Message: "success", Data: data}
}

// SuccessWithMessage 成功响应（带自定义消息）
// SuccessWithMessage 成功响应（自定义消息, 业务码=0）
func SuccessWithMessage(message string, data interface{}) *APIResponse {
	return &APIResponse{Code: 0, Message: message, Data: data}
}

// Error 错误响应
func Error(code int, message string) *APIResponse {
	return &APIResponse{
		Code:    code,
		Message: message,
	}
}

// BadRequest 400错误
func BadRequest(message string) *APIResponse {
	return &APIResponse{
		Code:    400,
		Message: message,
	}
}

// Unauthorized 401错误
func Unauthorized(message string) *APIResponse {
	return &APIResponse{
		Code:    401,
		Message: message,
	}
}

// Forbidden 403错误
func Forbidden(message string) *APIResponse {
	return &APIResponse{
		Code:    403,
		Message: message,
	}
}

// NotFound 404错误
func NotFound(message string) *APIResponse {
	return &APIResponse{
		Code:    404,
		Message: message,
	}
}

// InternalServerError 500错误
func InternalServerError(message string) *APIResponse {
	return &APIResponse{
		Code:    500,
		Message: message,
	}
}
