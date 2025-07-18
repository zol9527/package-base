package storages

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"package-base/internal/core/interfaces"
	"path/filepath"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/pkg/errors"
)

// FileConfig 文件存储配置结构
//
// 描述: 本地文件系统存储配置
type FileConfig struct {
	RootPath      string            `json:"root_path"`      // 根路径
	Permission    os.FileMode       `json:"permission"`     // 文件权限
	DirPermission os.FileMode       `json:"dir_permission"` // 目录权限
	MaxFileSize   int64             `json:"max_file_size"`  // 最大文件大小（字节）
	AllowedExts   []string          `json:"allowed_exts"`   // 允许的文件扩展名
	Metadata      map[string]string `json:"metadata"`       // 默认元数据
	BaseURL       string            `json:"base_url"`       // 基础URL（用于生成访问链接）
}

// GetType 获取存储类型
//
// 描述: 返回存储类型标识
//
// 返回:
//
//	string: 存储类型
func (c *FileConfig) GetType() string {
	return "file"
}

// Validate 验证配置
//
// 描述: 验证文件存储配置的有效性
//
// 返回:
//
//	error: 验证错误信息
func (c *FileConfig) Validate() error {
	if c.RootPath == "" {
		return errors.New("文件存储根路径不能为空")
	}

	// 检查路径是否存在，不存在则创建
	if !fileutil.IsExist(c.RootPath) {
		if err := os.MkdirAll(c.RootPath, 0755); err != nil {
			return errors.Wrap(err, "创建根路径失败")
		}
	}

	// 检查路径是否为目录
	if info, err := os.Stat(c.RootPath); err != nil {
		return errors.Wrap(err, "获取根路径信息失败")
	} else if !info.IsDir() {
		return errors.New("根路径必须是目录")
	}

	// 设置默认值
	if c.Permission == 0 {
		c.Permission = 0644
	}
	if c.DirPermission == 0 {
		c.DirPermission = 0755
	}
	if c.MaxFileSize <= 0 {
		c.MaxFileSize = 100 * 1024 * 1024 // 默认100MB
	}
	if c.Metadata == nil {
		c.Metadata = make(map[string]string)
	}

	return nil
}

// GetConnectionString 获取连接字符串
//
// 描述: 返回文件存储根路径
//
// 返回:
//
//	string: 连接字符串
func (c *FileConfig) GetConnectionString() string {
	return c.RootPath
}

// FileStorage 文件存储实现
//
// 描述: 基于本地文件系统的文件存储实现
type FileStorage struct {
	config *FileConfig
	logger interfaces.Logger
}

// NewFileStorage 创建文件存储实例
//
// 描述: 根据配置创建文件存储实例
//
// 参数:
//
//	config (*FileConfig): 文件存储配置
//	loggerInstance (root_interfaces.Logger): 日志记录器
//
// 返回:
//
//	interfaces.FileStorage: 文件存储接口
//	error: 错误信息
func NewFileStorage(config *FileConfig, loggerInstance interfaces.Logger) (interfaces.FileStorage, error) {
	if config == nil {
		return nil, errors.New("文件存储配置不能为空")
	}
	if loggerInstance == nil {
		return nil, errors.New("日志记录器不能为空")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "文件存储配置验证失败")
	}

	storage := &FileStorage{
		config: config,
		logger: loggerInstance.With(interfaces.String("component", "file-storage")),
	}
	storage.logger.Info("本地文件存储已初始化", interfaces.String("component", "file-storage"), interfaces.String("root_path", config.RootPath))
	return storage, nil
}

// Connect 连接到存储
//
// 描述: 文件存储不需要建立连接，检查根路径可用性
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Connect(ctx context.Context) error {
	s.logger.Debug("正在检查本地文件存储根路径...", interfaces.String("component", "file-storage"), interfaces.String("path", s.config.RootPath))
	// 检查根路径是否可访问
	if !fileutil.IsExist(s.config.RootPath) {
		err := errors.New("文件存储根路径不存在")
		s.logger.Error("本地文件存储根路径检查失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(err))
		return err
	}

	// 检查读写权限
	testFile := filepath.Join(s.config.RootPath, ".test_access")
	if err := os.WriteFile(testFile, []byte("test"), s.config.Permission); err != nil {
		wrappedErr := errors.Wrap(err, "文件存储根路径没有写权限")
		s.logger.Error("本地文件存储根路径没有写权限", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	// 清理测试文件
	os.Remove(testFile)

	s.logger.Debug("本地文件存储根路径可用", interfaces.String("component", "file-storage"))
	return nil
}

// Close 关闭存储连接
//
// 描述: 文件存储不需要关闭连接
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Close() error {
	s.logger.Debug("关闭本地文件存储连接 (无操作)", interfaces.String("component", "file-storage"))
	return nil
}

// Ping 测试连接状态
//
// 描述: 测试文件存储是否可用
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Ping(ctx context.Context) error {
	s.logger.Debug("Pinging 本地文件存储...", interfaces.String("component", "file-storage"))
	if err := s.Connect(ctx); err != nil {
		// Connect 方法内部已经记录了错误，这里只记录 Ping 失败的结果
		s.logger.Error("Ping 本地文件存储失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(err))
		return err
	}
	s.logger.Debug("Ping 本地文件存储成功", interfaces.String("component", "file-storage"))
	return nil
}

// IsHealthy 检查存储健康状态
//
// 描述: 检查文件存储是否健康
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	bool: 是否健康
func (s *FileStorage) IsHealthy(ctx context.Context) bool {
	s.logger.Debug("正在检查本地文件存储健康状态...", interfaces.String("component", "file-storage"))
	isHealthy := s.Ping(ctx) == nil
	s.logger.Debug("本地文件存储健康状态", interfaces.String("component", "file-storage"), interfaces.Bool("is_healthy", isHealthy))
	return isHealthy
}

// getFullPath 获取完整路径
//
// 描述: 根据相对路径获取完整的文件系统路径
//
// 参数:
//
//	path (string): 相对路径
//
// 返回:
//
//	string: 完整路径
func (s *FileStorage) getFullPath(path string) string {
	// 如果传入的是绝对路径并且就在 RootPath 之下，则直接返回
	if filepath.IsAbs(path) {
		cleanPath := filepath.Clean(path)
		rootClean := filepath.Clean(s.config.RootPath)
		if strings.HasPrefix(cleanPath, rootClean) {
			return cleanPath
		}
	}

	// 如果传入的相对路径已经包含 RootPath 前缀，直接返回清理后的相对路径，避免重复拼接
	cleanPath := filepath.Clean(path)
	rootClean := filepath.Clean(s.config.RootPath)
	if strings.HasPrefix(cleanPath, rootClean+string(os.PathSeparator)) || cleanPath == rootClean {
		return cleanPath
	}

	// 否则将相对路径拼接到 RootPath，防止目录遍历攻击
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	return filepath.Join(s.config.RootPath, cleanPath)
}

// validateFileExtension 验证文件扩展名
//
// 描述: 检查文件扩展名是否在允许列表中
//
// 参数:
//
//	path (string): 文件路径
//
// 返回:
//
//	error: 验证错误
func (s *FileStorage) validateFileExtension(path string) error {
	if len(s.config.AllowedExts) == 0 {
		return nil // 没有限制
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return errors.New("文件必须有扩展名")
	}

	for _, allowedExt := range s.config.AllowedExts {
		if strings.ToLower(allowedExt) == ext {
			return nil
		}
	}

	return errors.Errorf("不允许的文件类型: %s", ext)
}

// Upload 上传文件
//
// 描述: 上传文件到指定路径
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	data ([]byte): 文件数据
//	metadata (map[string]string): 文件元数据
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Upload(ctx context.Context, path string, data []byte, metadata map[string]string) error {
	s.logger.Debug("正在上传文件", interfaces.String("component", "file-storage"), interfaces.String("path", path), interfaces.Int("size", len(data)))
	if path == "" {
		return errors.New("文件路径不能为空")
	}

	if int64(len(data)) > s.config.MaxFileSize {
		return errors.Errorf("文件大小超过限制: %d > %d", len(data), s.config.MaxFileSize)
	}

	// 验证文件扩展名
	if err := s.validateFileExtension(path); err != nil {
		s.logger.Error("文件类型不被允许", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("path", path))
		return err
	}

	fullPath := s.getFullPath(path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, s.config.DirPermission); err != nil {
		wrappedErr := errors.Wrap(err, "创建目录失败")
		s.logger.Error("创建目录失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("dir", dir))
		return wrappedErr
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, s.config.Permission); err != nil {
		wrappedErr := errors.Wrap(err, "写入文件失败")
		s.logger.Error("写入文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
		return wrappedErr
	}

	// 保存元数据
	if len(metadata) > 0 {
		if err := s.saveMetadata(fullPath, metadata); err != nil {
			// 删除已上传的文件
			os.Remove(fullPath)
			s.logger.Error("保存元数据失败，已回滚上传的文件", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("path", fullPath))
			return errors.Wrap(err, "保存元数据失败")
		}
	}

	s.logger.Debug("文件上传成功", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
	return nil
}

// Download 下载文件
//
// 描述: 从指定路径下载文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	[]byte: 文件数据
//	error: 错误信息
func (s *FileStorage) Download(ctx context.Context, path string) ([]byte, error) {
	s.logger.Debug("正在下载文件", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return nil, errors.New("文件路径不能为空")
	}

	fullPath := s.getFullPath(path)

	// 检查文件是否存在
	if !fileutil.IsExist(fullPath) {
		err := errors.New("文件不存在")
		s.logger.Warn("下载失败，文件不存在", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
		return nil, err
	}

	// 检查是否为文件
	if info, err := os.Stat(fullPath); err != nil {
		wrappedErr := errors.Wrap(err, "获取文件信息失败")
		s.logger.Error("获取文件信息失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
		return nil, wrappedErr
	} else if info.IsDir() {
		err := errors.New("路径是目录，不是文件")
		s.logger.Error("无法下载目录", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("path", fullPath))
		return nil, err
	}

	// 读取文件
	data, err := os.ReadFile(fullPath)
	if err != nil {
		wrappedErr := errors.Wrap(err, "读取文件失败")
		s.logger.Error("读取文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
		return nil, wrappedErr
	}

	s.logger.Debug("文件下载成功", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath), interfaces.Int("size", len(data)))
	return data, nil
}

// Delete 删除文件
//
// 描述: 删除指定路径的文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Delete(ctx context.Context, path string) error {
	s.logger.Debug("准备删除文件", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return errors.New("文件路径不能为空")
	}

	fullPath := s.getFullPath(path)
	s.logger.Debug("  - 完整路径", interfaces.String("component", "file-storage"), interfaces.String("fullPath", fullPath))

	// 检查文件是否存在
	if !fileutil.IsExist(fullPath) {
		s.logger.Warn("文件删除被跳过，因为文件不存在", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
		return nil // 文件不存在，认为删除成功
	}

	// 删除元数据文件
	s.deleteMetadata(fullPath)

	// 删除文件
	if err := os.Remove(fullPath); err != nil {
		wrappedErr := errors.Wrap(err, "删除文件失败")
		s.logger.Error("删除文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
		return wrappedErr
	}

	s.logger.Debug("文件删除成功", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
	return nil
}

// Exists 检查文件是否存在
//
// 描述: 检查指定路径的文件是否存在
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	bool: 是否存在
//	error: 错误信息
func (s *FileStorage) Exists(ctx context.Context, path string) (bool, error) {
	s.logger.Debug("正在检查文件是否存在", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return false, errors.New("文件路径不能为空")
	}

	fullPath := s.getFullPath(path)
	exists := fileutil.IsExist(fullPath)

	if exists {
		// 确保是文件而不是目录
		if info, err := os.Stat(fullPath); err != nil {
			wrappedErr := errors.Wrap(err, "获取文件信息失败")
			s.logger.Error("检查文件存在性时获取信息失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
			return false, wrappedErr
		} else if info.IsDir() {
			s.logger.Debug("  - 路径存在，但是一个目录", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
			return false, nil
		}
	}

	s.logger.Debug("文件存在性检查完成", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath), interfaces.Bool("exists", exists))
	return exists, nil
}

// GetMetadata 获取文件元数据
//
// 描述: 获取文件的元数据信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	map[string]string: 元数据映射
//	error: 错误信息
func (s *FileStorage) GetMetadata(ctx context.Context, path string) (map[string]string, error) {
	s.logger.Debug("正在获取文件元数据", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return nil, errors.New("文件路径不能为空")
	}

	fullPath := s.getFullPath(path)

	// 检查文件是否存在
	if !fileutil.IsExist(fullPath) {
		err := errors.New("文件不存在")
		s.logger.Warn("获取元数据失败，文件不存在", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
		return nil, err
	}

	// 获取基本信息
	info, err := os.Stat(fullPath)
	if err != nil {
		wrappedErr := errors.Wrap(err, "获取文件信息失败")
		s.logger.Error("获取元数据时获取文件信息失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", fullPath))
		return nil, wrappedErr
	}

	metadata := make(map[string]string)

	// 基本元数据
	metadata["size"] = fmt.Sprintf("%d", info.Size())
	metadata["mod_time"] = info.ModTime().Format(time.RFC3339)
	metadata["content_type"] = s.getContentType(path)

	// 读取自定义元数据
	customMetadata, err := s.loadMetadata(fullPath)
	if err == nil {
		for k, v := range customMetadata {
			metadata[k] = v
		}
	}

	s.logger.Debug("获取文件元数据成功", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	return metadata, nil
}

// SetMetadata 设置文件元数据
//
// 描述: 设置文件的元数据信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	metadata (map[string]string): 元数据映射
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) SetMetadata(ctx context.Context, path string, metadata map[string]string) error {
	s.logger.Debug("正在设置文件元数据", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return errors.New("文件路径不能为空")
	}

	fullPath := s.getFullPath(path)

	// 检查文件是否存在
	if !fileutil.IsExist(fullPath) {
		err := errors.New("文件不存在")
		s.logger.Warn("设置元数据失败，文件不存在", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
		return err
	}

	// 保存元数据
	err := s.saveMetadata(fullPath, metadata)
	if err != nil {
		s.logger.Error("设置元数据失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("path", fullPath))
		return err
	}

	s.logger.Debug("设置文件元数据成功", interfaces.String("component", "file-storage"), interfaces.String("path", fullPath))
	return nil
}

// List 列出文件
//
// 描述: 列出指定前缀下的所有文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	prefix (string): 路径前缀
//
// 返回:
//
//	[]interfaces.FileInfo: 文件信息列表
//	error: 错误信息
func (s *FileStorage) List(ctx context.Context, prefix string) ([]interfaces.FileInfo, error) {
	s.logger.Debug("正在列出文件", interfaces.String("component", "file-storage"), interfaces.String("prefix", prefix))
	fullPrefix := s.getFullPath(prefix)
	var fileInfos []interfaces.FileInfo

	// 如果前缀是文件，直接返回该文件信息
	if fileutil.IsExist(fullPrefix) {
		if info, err := os.Stat(fullPrefix); err == nil && !info.IsDir() {
			fileInfo, err := s.buildFileInfo(prefix, info)
			if err != nil {
				s.logger.Error("构建文件信息失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("prefix", prefix))
				return nil, err
			}
			return []interfaces.FileInfo{*fileInfo}, nil
		}
	}

	// 遍历目录
	err := filepath.WalkDir(fullPrefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续遍历
		}

		// 获取相对路径
		relPath, err := filepath.Rel(s.config.RootPath, path)
		if err != nil {
			return nil
		}

		// 转换为Unix风格路径
		relPath = filepath.ToSlash(relPath)

		// 获取文件信息
		info, err := d.Info()
		if err != nil {
			return nil
		}

		fileInfo, err := s.buildFileInfo(relPath, info)
		if err != nil {
			return nil
		}

		fileInfos = append(fileInfos, *fileInfo)
		return nil
	})

	if err != nil {
		s.logger.Error("遍历目录失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("prefix", fullPrefix))
		return nil, errors.Wrap(err, "遍历目录失败")
	}

	s.logger.Debug("列出文件成功", interfaces.String("component", "file-storage"), interfaces.String("prefix", prefix), interfaces.Int("count", len(fileInfos)))
	return fileInfos, nil
}

// Copy 复制文件
//
// 描述: 复制文件到新路径
//
// 参数:
//
//	ctx (context.Context): 上下文
//	srcPath (string): 源文件路径
//	destPath (string): 目标文件路径
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Copy(ctx context.Context, srcPath, destPath string) error {
	s.logger.Debug("正在复制文件", interfaces.String("component", "file-storage"), interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	if srcPath == "" || destPath == "" {
		return errors.New("源路径和目标路径不能为空")
	}

	// 读取源文件
	data, err := s.Download(ctx, srcPath)
	if err != nil {
		// Download 内部已记录日志
		return errors.Wrap(err, "读取源文件失败")
	}

	// 获取源文件元数据
	metadata, err := s.GetMetadata(ctx, srcPath)
	if err != nil {
		s.logger.Warn("复制文件时获取源文件元数据失败，将不带元数据复制", interfaces.String("component", "file-storage"), interfaces.ErrorField(err), interfaces.String("source", srcPath))
		metadata = make(map[string]string) // 忽略元数据错误
	}

	// 写入目标文件
	if err := s.Upload(ctx, destPath, data, metadata); err != nil {
		// Upload 内部已记录日志
		return errors.Wrap(err, "写入目标文件失败")
	}

	s.logger.Debug("复制文件成功", interfaces.String("component", "file-storage"), interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	return nil
}

// Move 移动文件
//
// 描述: 移动文件到新路径
//
// 参数:
//
//	ctx (context.Context): 上下文
//	srcPath (string): 源文件路径
//	destPath (string): 目标文件路径
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) Move(ctx context.Context, srcPath, destPath string) error {
	s.logger.Debug("正在移动文件", interfaces.String("component", "file-storage"), interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	if srcPath == "" || destPath == "" {
		return errors.New("源路径和目标路径不能为空")
	}

	srcFullPath := s.getFullPath(srcPath)
	destFullPath := s.getFullPath(destPath)

	// 确保目标目录存在
	destDir := filepath.Dir(destFullPath)
	if err := os.MkdirAll(destDir, s.config.DirPermission); err != nil {
		wrappedErr := errors.Wrap(err, "创建目标目录失败")
		s.logger.Error("移动文件时创建目标目录失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("destination_dir", destDir))
		return wrappedErr
	}

	// 移动文件
	if err := os.Rename(srcFullPath, destFullPath); err != nil {
		s.logger.Warn("rename 移动文件失败，尝试 copy-then-delete 策略", interfaces.String("component", "file-storage"), interfaces.ErrorField(err))
		// 如果rename失败，尝试复制然后删除
		if err := s.Copy(ctx, srcPath, destPath); err != nil {
			// Copy 内部已记录日志
			return errors.Wrap(err, "复制文件失败")
		}
		if err := s.Delete(ctx, srcPath); err != nil {
			// Delete 内部已记录日志
			return errors.Wrap(err, "删除源文件失败")
		}
	}

	// 移动元数据文件
	s.moveMetadata(srcFullPath, destFullPath)

	s.logger.Debug("移动文件成功", interfaces.String("component", "file-storage"), interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	return nil
}

// GetURL 获取文件访问URL
//
// 描述: 获取文件的访问URL
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	expiration (time.Duration): 过期时间（文件存储忽略）
//
// 返回:
//
//	string: 访问URL
//	error: 错误信息
func (s *FileStorage) GetURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	s.logger.Debug("正在获取文件URL", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	if path == "" {
		return "", errors.New("文件路径不能为空")
	}

	// 检查文件是否存在
	exists, err := s.Exists(ctx, path)
	if err != nil {
		// Exists 内部已记录日志
		return "", err
	}
	if !exists {
		err := errors.New("文件不存在")
		s.logger.Warn("获取URL失败，文件不存在", interfaces.String("component", "file-storage"), interfaces.String("path", path))
		return "", err
	}

	if s.config.BaseURL == "" {
		err := errors.New("未配置基础URL")
		s.logger.Error("获取URL失败，未配置基础URL", interfaces.String("component", "file-storage"), interfaces.ErrorField(err))
		return "", err
	}

	// 构建URL
	url := fmt.Sprintf("%s/%s", strings.TrimRight(s.config.BaseURL, "/"), strings.TrimLeft(path, "/"))
	s.logger.Debug("获取文件URL成功", interfaces.String("component", "file-storage"), interfaces.String("url", url))
	return url, nil
}

// saveMetadata 保存元数据
//
// 描述: 将元数据保存到文件
//
// 参数:
//
//	filePath (string): 文件路径
//	metadata (map[string]string): 元数据
//
// 返回:
//
//	error: 错误信息
func (s *FileStorage) saveMetadata(filePath string, metadata map[string]string) error {
	s.logger.Debug("正在保存元数据", interfaces.String("component", "file-storage"), interfaces.String("path", filePath))
	metadataPath := filePath + ".meta"

	// 合并默认元数据
	finalMetadata := make(map[string]string)
	for k, v := range s.config.Metadata {
		finalMetadata[k] = v
	}
	for k, v := range metadata {
		finalMetadata[k] = v
	}

	data, err := json.Marshal(finalMetadata)
	if err != nil {
		wrappedErr := errors.Wrap(err, "序列化元数据失败")
		s.logger.Error("序列化元数据失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", filePath))
		return wrappedErr
	}

	if err := os.WriteFile(metadataPath, data, s.config.Permission); err != nil {
		wrappedErr := errors.Wrap(err, "写入元数据文件失败")
		s.logger.Error("写入元数据文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", metadataPath))
		return wrappedErr
	}
	s.logger.Debug("元数据保存成功", interfaces.String("component", "file-storage"), interfaces.String("path", metadataPath))
	return nil
}

// loadMetadata 加载元数据
//
// 描述: 从文件加载元数据
//
// 参数:
//
//	filePath (string): 文件路径
//
// 返回:
//
//	map[string]string: 元数据
//	error: 错误信息
func (s *FileStorage) loadMetadata(filePath string) (map[string]string, error) {
	s.logger.Debug("📥 正在加载元数据", interfaces.String("component", "file-storage"), interfaces.String("path", filePath))
	metadataPath := filePath + ".meta"

	if !fileutil.IsExist(metadataPath) {
		s.logger.Debug("元数据文件不存在，返回空map", interfaces.String("component", "file-storage"), interfaces.String("path", metadataPath))
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		wrappedErr := errors.Wrap(err, "读取元数据文件失败")
		s.logger.Error("读取元数据文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", metadataPath))
		return nil, wrappedErr
	}

	var metadata map[string]string
	if err := json.Unmarshal(data, &metadata); err != nil {
		wrappedErr := errors.Wrap(err, "解析元数据失败")
		s.logger.Error("解析元数据失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", metadataPath))
		return nil, wrappedErr
	}

	s.logger.Debug("元数据加载成功", interfaces.String("component", "file-storage"), interfaces.String("path", metadataPath))
	return metadata, nil
}

// deleteMetadata 删除元数据
//
// 描述: 删除元数据文件
//
// 参数:
//
//	filePath (string): 文件路径
func (s *FileStorage) deleteMetadata(filePath string) {
	metadataPath := filePath + ".meta"
	s.logger.Debug("🧹 正在删除元数据文件", interfaces.String("component", "file-storage"), interfaces.String("path", metadataPath))
	os.Remove(metadataPath) // 忽略错误
}

// moveMetadata 移动元数据
//
// 描述: 移动元数据文件
//
// 参数:
//
//	srcPath (string): 源文件路径
//	destPath (string): 目标文件路径
func (s *FileStorage) moveMetadata(srcPath, destPath string) {
	srcMetadataPath := srcPath + ".meta"
	destMetadataPath := destPath + ".meta"
	s.logger.Debug("🚚 正在移动元数据文件", interfaces.String("component", "file-storage"), interfaces.String("source", srcMetadataPath), interfaces.String("destination", destMetadataPath))

	if fileutil.IsExist(srcMetadataPath) {
		os.Rename(srcMetadataPath, destMetadataPath) // 忽略错误
	}
}

// getContentType 获取内容类型
//
// 描述: 根据文件扩展名推断内容类型
//
// 参数:
//
//	path (string): 文件路径
//
// 返回:
//
//	string: 内容类型
func (s *FileStorage) getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".avi":  "video/avi",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// buildFileInfo 构建文件信息
//
// 描述: 根据os.FileInfo构建interfaces.FileInfo
//
// 参数:
//
//	path (string): 文件路径
//	info (os.FileInfo): 文件信息
//
// 返回:
//
//	*interfaces.FileInfo: 文件信息对象
//	error: 错误信息
func (s *FileStorage) buildFileInfo(path string, info os.FileInfo) (*interfaces.FileInfo, error) {
	// 加载元数据
	fullPath := s.getFullPath(path)
	metadata, err := s.loadMetadata(fullPath)
	if err != nil {
		metadata = make(map[string]string)
	}

	fileInfo := &interfaces.FileInfo{
		Path:         path,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		IsDir:        info.IsDir(),
		ContentType:  s.getContentType(path),
		Metadata:     metadata,
		ETag:         fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix()),
		LastModified: info.ModTime(),
	}

	return fileInfo, nil
}

// GetConfig 获取配置
//
// 描述: 获取文件存储配置信息
//
// 返回:
//
//	*FileConfig: 文件存储配置
func (s *FileStorage) GetConfig() *FileConfig {
	return s.config
}

// Save 保存文件到本地
func (s *FileStorage) Save(file *multipart.FileHeader) (string, error) {
	s.logger.Debug("准备保存上传的文件", interfaces.String("component", "file-storage"), interfaces.String("original_filename", file.Filename), interfaces.Int64("size", file.Size))
	// 在启动时已通过 Connect 方法进行了健康检查，此处不再重复检查
	// if !s.IsHealthy(context.Background()) {
	// 	return "", errors.New("本地存储未连接")
	// }

	// 生成唯一文件名
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		wrappedErr := errors.Wrap(err, "生成随机文件名失败")
		s.logger.Error("生成随机文件名失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr))
		return "", wrappedErr
	}
	fileName := hex.EncodeToString(randBytes) + filepath.Ext(file.Filename)
	filePath := filepath.Join(s.config.RootPath, fileName)
	s.logger.Debug("  - 文件将保存至", interfaces.String("component", "file-storage"), interfaces.String("path", filePath))

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		wrappedErr := errors.Wrap(err, "创建目标文件失败")
		s.logger.Error("创建目标文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr), interfaces.String("path", filePath))
		return "", wrappedErr
	}
	defer dst.Close()

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		wrappedErr := errors.Wrap(err, "打开源文件失败")
		s.logger.Error("打开源文件失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr))
		return "", wrappedErr
	}
	defer src.Close()

	// 拷贝文件
	_, err = io.Copy(dst, src)
	if err != nil {
		wrappedErr := errors.Wrap(err, "拷贝文件内容失败")
		s.logger.Error("拷贝文件内容失败", interfaces.String("component", "file-storage"), interfaces.ErrorField(wrappedErr))
		return "", wrappedErr
	}

	s.logger.Debug("上传的文件保存成功", interfaces.String("component", "file-storage"), interfaces.String("path", filePath))
	return filePath, nil
}

// Driver 返回驱动名称
func (s *FileStorage) Driver() string {
	return "local"
}

// Get an opened file
func (s *FileStorage) Get(ctx context.Context, path string) (*os.File, error) {
	s.logger.Debug("正在打开文件", interfaces.String("component", "file-storage"), interfaces.String("path", path))
	return os.Open(path)
}
