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

// RoleCreateRequest 角色创建请求
type RoleCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// RoleResponse 角色响应数据
type RoleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BuiltIn     bool   `json:"built_in"`
	Enabled     bool   `json:"enabled"`
}

// UserBindRolesRequest 用户绑定角色请求
type UserBindRolesRequest struct {
	RoleIDs       []uint  `json:"role_ids"`
	Scope         *string `json:"scope,omitempty"`
	ExplicitEmpty *bool   `json:"explicit_empty,omitempty"` // 为 true 且 role_ids 为空时，表示解除该 scope 下所有绑定
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
