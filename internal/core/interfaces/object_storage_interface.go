package interfaces

import (
	"context"
	"io"
	"time"
)

// Function: ObjectStorage
//
// Description: 对象存储通用接口（统一本地/远程最小抽象），提供基础对象操作与预签名能力，供业务逻辑按最小依赖使用。
//
// Parameters:
//
//	无（接口定义）
//
// Returns:
//
//	无（接口定义）
type ObjectStorage interface {
	// Put 存储对象（若已存在则覆盖）
	//
	// Parameters:
	// 	ctx(context.Context): 上下文
	// 	path(string): 对象路径（逻辑 Key）
	// 	r(io.Reader): 数据流
	// 	metadata(map[string]string): 可选元数据
	// Returns:
	// 	error(error): 失败时返回错误
	Put(ctx context.Context, path string, r io.Reader, metadata map[string]string) error

	// Get 读取对象至 writer
	//
	// Parameters:
	// 	ctx(context.Context): 上下文
	// 	path(string): 对象路径
	// 	w(io.Writer): 输出 writer
	// Returns:
	// 	error(error): 失败时返回错误
	Get(ctx context.Context, path string, w io.Writer) error

	// Delete 删除对象
	//
	// Parameters:
	// 	ctx(context.Context): 上下文
	// 	path(string): 对象路径
	// Returns:
	// 	error(error): 失败时返回错误
	Delete(ctx context.Context, path string) error

	// Exists 判断对象是否存在
	//
	// Parameters:
	// 	ctx(context.Context): 上下文
	// 	path(string): 对象路径
	// Returns:
	// 	exists(bool): 是否存在
	// 	error(error): 失败时返回错误
	Exists(ctx context.Context, path string) (bool, error)

	// Presign 生成预签名 URL（上传/下载统一）
	//
	// Parameters:
	// 	ctx(context.Context): 上下文
	// 	path(string): 对象路径
	// 	method(string): HTTP 方法 (GET/PUT)
	// 	expiration(time.Duration): 过期时间
	// 	metadata(map[string]string): 需要参与签名的元数据（可选）
	// Returns:
	// 	url(string): 预签名地址
	// 	error(error): 失败时返回错误
	Presign(ctx context.Context, path string, method string, expiration time.Duration, metadata map[string]string) (string, error)

	// Driver 返回底层驱动标识
	Driver() string
}

// Function: ObjectStorageError
//
// Description: 自定义对象存储错误，统一错误码便于上层解析。
//
// Parameters:
//
//	无（结构体定义）
//
// Returns:
//
//	无（结构体定义）
type ObjectStorageError struct {
	Code    string // 错误码 (e.g. not_found, permission_denied, internal, invalid_argument)
	Message string // 人类可读描述
	Err     error  // 底层错误
}

// Function: Error
//
// Description: 实现 error 接口，返回格式化后的错误信息。
//
// Parameters:
//
//	无
//
// Returns:
//
//	string(string): 错误字符串
func (e *ObjectStorageError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Function: Unwrap
//
// Description: 返回底层错误以支持 errors.Is / errors.As。
//
// Parameters:
//
//	无
//
// Returns:
//
//	error(error): 底层错误
func (e *ObjectStorageError) Unwrap() error { return e.Err }
