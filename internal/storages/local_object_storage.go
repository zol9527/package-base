package storages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"package-base/internal/core/interfaces"
)

// Function: LocalObjectStorage
//
// Description: 基于本地文件系统的对象存储实现，满足 ObjectStorage 最小能力，便于开发/测试环境使用。
//
// Parameters:
//
//	无（结构体定义）
//
// Returns:
//
//	无（结构体定义）
type LocalObjectStorage struct {
	baseDir string
}

// Function: NewLocalObjectStorage
//
// Description: 创建 LocalObjectStorage 实例，若目录不存在会尝试创建。
//
// Parameters:
//
//	baseDir(string): 基础目录
//
// Returns:
//
//	storage(*LocalObjectStorage): 实例
//	err(error): 错误
func NewLocalObjectStorage(baseDir string) (*LocalObjectStorage, error) {
	if baseDir == "" {
		baseDir = "data/objects"
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil { // #nosec G301
		return nil, fmt.Errorf("创建对象存储目录失败: %w", err)
	}
	return &LocalObjectStorage{baseDir: baseDir}, nil
}

// Function: Put
//
// Description: 写入对象（覆盖式）。
//
// Parameters:
//
//	ctx(context.Context): 上下文（当前未使用，可用于 future 取消）
//	path(string): 逻辑路径
//	r(io.Reader): 数据源
//	metadata(map[string]string): 元数据（本地实现忽略，可扩展写 sidecar）
//
// Returns:
//
//	err(error): 错误
func (s *LocalObjectStorage) Put(ctx context.Context, path string, r io.Reader, metadata map[string]string) error { // nolint:revive // 语义保持
	if strings.HasPrefix(path, "..") {
		return errors.New("非法路径")
	}
	fullPath := filepath.Join(s.baseDir, filepath.Clean(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil { // #nosec G301
		return fmt.Errorf("创建目录失败: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// Function: Get
//
// Description: 读取对象内容写入 writer。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	path(string): 路径
//	w(io.Writer): 目标 writer
//
// Returns:
//
//	err(error): 错误
func (s *LocalObjectStorage) Get(ctx context.Context, path string, w io.Writer) error { // nolint:revive
	fullPath := filepath.Join(s.baseDir, filepath.Clean(path))
	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &interfaces.ObjectStorageError{Code: "not_found", Message: "对象不存在", Err: err}
		}
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	return nil
}

// Function: Delete
//
// Description: 删除对象。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	path(string): 路径
//
// Returns:
//
//	err(error): 错误
func (s *LocalObjectStorage) Delete(ctx context.Context, path string) error { // nolint:revive
	fullPath := filepath.Join(s.baseDir, filepath.Clean(path))
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 幂等
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// Function: Exists
//
// Description: 判断对象是否存在。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	path(string): 路径
//
// Returns:
//
//	exists(bool): 是否存在
//	err(error): 错误
func (s *LocalObjectStorage) Exists(ctx context.Context, path string) (bool, error) { // nolint:revive
	fullPath := filepath.Join(s.baseDir, filepath.Clean(path))
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat 失败: %w", err)
}

// Function: Presign
//
// Description: 生成伪预签名 URL（本地实现返回文件相对路径 + 过期时间戳查询参数，仅供测试）。
//
// Parameters:
//
//	ctx(context.Context): 上下文
//	path(string): 路径
//	method(string): HTTP 方法
//	expiration(time.Duration): 过期时长
//	metadata(map[string]string): 元数据（忽略）
//
// Returns:
//
//	url(string): 伪 URL
//	err(error): 错误
func (s *LocalObjectStorage) Presign(ctx context.Context, path string, method string, expiration time.Duration, metadata map[string]string) (string, error) { // nolint:revive
	if expiration <= 0 {
		expiration = 15 * time.Minute
	}
	expireAt := time.Now().Add(expiration).Unix()
	return fmt.Sprintf("local://%s?method=%s&exp=%d", filepath.Clean(path), method, expireAt), nil
}

// Function: Driver
//
// Description: 返回驱动名称。
//
// Parameters:
//
//	无
//
// Returns:
//
//	string(string): 驱动标识
func (s *LocalObjectStorage) Driver() string { return "local" }
