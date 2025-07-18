package storages

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"package-base/internal/core/interfaces"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"
)

// AliOSSConfig 阿里云OSS配置
//
// 描述: 阿里云OSS存储的配置参数
type AliOSSConfig struct {
	Endpoint        string `json:"endpoint"`          // OSS服务端点
	AccessKeyID     string `json:"access_key_id"`     // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret"` // 访问密钥
	BucketName      string `json:"bucket_name"`       // 存储桶名称
	BucketDomain    string `json:"bucket_domain"`     // 存储桶域名
	BasePath        string `json:"base_path"`         // 基础路径前缀
	Region          string `json:"region"`            // 区域
}

// GetType 获取存储类型
//
// 描述: 返回存储类型标识
//
// 返回:
//
//	string: 存储类型 "ali_oss"
func (c *AliOSSConfig) GetType() string {
	return "ali_oss"
}

// Validate 验证配置
//
// 描述: 验证阿里云OSS配置的完整性和有效性
//
// 返回:
//
//	error: 验证错误信息
func (c *AliOSSConfig) Validate() error {
	if c.Endpoint == "" {
		return errors.New("阿里云OSS endpoint 不能为空")
	}
	if c.AccessKeyID == "" {
		return errors.New("阿里云OSS access key ID 不能为空")
	}
	if c.AccessKeySecret == "" {
		return errors.New("阿里云OSS access key secret 不能为空")
	}
	if c.BucketName == "" {
		return errors.New("阿里云OSS bucket name 不能为空")
	}
	return nil
}

// GetConnectionString 获取连接字符串
//
// 描述: 返回阿里云OSS连接信息的字符串表示
//
// 返回:
//
//	string: 连接字符串
func (c *AliOSSConfig) GetConnectionString() string {
	return fmt.Sprintf("ali_oss://%s@%s/%s", c.AccessKeyID, c.Endpoint, c.BucketName)
}

// AliOSSStorage 阿里云OSS存储实现
//
// 描述: 基于阿里云OSS的文件存储实现
type AliOSSStorage struct {
	config    *AliOSSConfig
	client    *oss.Client
	bucket    *oss.Bucket
	connected bool
	logger    interfaces.Logger
}

// NewAliOSSStorage 创建阿里云OSS存储实例
//
// 描述: 创建并初始化阿里云OSS存储实例
//
// 参数:
//
//	config (*AliOSSConfig): 阿里云OSS配置
//	loggerInstance (root_interfaces.Logger): 日志记录器
//
// 返回:
//
//	(interfaces.RemoteStorage): 远程存储接口实例
//	(error): 错误信息
func NewAliOSSStorage(config *AliOSSConfig, loggerInstance interfaces.Logger) (interfaces.RemoteOssStorage, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "阿里云OSS配置验证失败")
	}

	storage := &AliOSSStorage{
		config:    config,
		connected: false,
		logger:    loggerInstance.With(interfaces.String("component", "ali-oss-storage")),
	}

	return storage, nil
}

// Connect 连接到阿里云OSS
//
// 描述: 初始化阿里云OSS客户端连接
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 连接错误信息
func (s *AliOSSStorage) Connect(ctx context.Context) error {
	s.logger.Debug("正在连接到阿里云OSS...")
	if s.connected {
		s.logger.Debug("连接已存在，跳过")
		return nil
	}

	// 创建OSS客户端
	client, err := oss.New(s.config.Endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		s.logger.Error("创建阿里云OSS客户端失败", interfaces.ErrorField(err))
		return errors.Wrap(err, "创建阿里云OSS客户端失败")
	}

	// 获取存储桶
	bucket, err := client.Bucket(s.config.BucketName)
	if err != nil {
		s.logger.Error("获取阿里云OSS存储桶失败", interfaces.ErrorField(err), interfaces.String("bucket_name", s.config.BucketName))
		return errors.Wrap(err, "获取阿里云OSS存储桶失败")
	}

	s.client = client
	s.bucket = bucket
	s.connected = true

	s.logger.Info("阿里云OSS存储连接成功")
	return nil
}

// Close 关闭阿里云OSS连接
//
// 描述: 关闭阿里云OSS客户端连接
//
// 返回:
//
//	error: 关闭错误信息
func (s *AliOSSStorage) Close() error {
	s.logger.Debug("正在关闭阿里云OSS连接...")
	if !s.connected {
		s.logger.Debug("连接不存在，无需关闭")
		return nil
	}

	s.client = nil
	s.bucket = nil
	s.connected = false

	s.logger.Info("阿里云OSS存储连接已关闭")
	return nil
}

// Ping 测试阿里云OSS连接
//
// 描述: 测试阿里云OSS服务连接状态
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 连接测试错误信息
func (s *AliOSSStorage) Ping(ctx context.Context) error {
	s.logger.Debug("Pinging 阿里云OSS...")
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("Ping失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	// 通过检查存储桶是否存在来测试连接
	// 使用一个不存在的对象名来测试，避免依赖特定对象
	_, err := s.bucket.IsObjectExist("health_check_test_object")
	if err != nil {
		// IsObjectExist 在对象不存在时不返回错误，所以任何错误都表示连接有问题
		wrappedErr := errors.Wrap(err, "阿里云OSS连接测试失败")
		s.logger.Error("Ping 阿里云OSS失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	s.logger.Debug("Ping 阿里云OSS成功")
	return nil
}

// IsHealthy 检查阿里云OSS健康状态
//
// 描述: 检查阿里云OSS服务的健康状态
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	bool: 是否健康
func (s *AliOSSStorage) IsHealthy(ctx context.Context) bool {
	s.logger.Debug("正在检查阿里云OSS健康状态...")
	isHealthy := s.Ping(ctx) == nil
	s.logger.Debug("阿里云OSS健康状态检查完成", interfaces.Bool("is_healthy", isHealthy))
	return isHealthy
}

// Upload 上传文件到阿里云OSS
//
// 描述: 将文件数据上传到阿里云OSS存储
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
//	error: 上传错误信息
func (s *AliOSSStorage) Upload(ctx context.Context, path string, data []byte, metadata map[string]string) error {
	s.logger.Debug("正在上传文件到阿里云OSS", interfaces.String("path", path), interfaces.Int("size", len(data)))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("上传文件失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	// 标准化路径
	normalizedPath := s.normalizePath(path)

	// 创建读取器
	reader := bytes.NewReader(data)

	// 准备上传选项
	var options []oss.Option

	// 设置内容类型
	if metadata != nil && metadata["content-type"] != "" {
		options = append(options, oss.ContentType(metadata["content-type"]))
	} else {
		contentType := s.getContentType(normalizedPath)
		options = append(options, oss.ContentType(contentType))
	}

	// 设置其他元数据
	for key, value := range metadata {
		if key != "content-type" {
			options = append(options, oss.Meta(key, value))
		}
	}

	// 上传文件
	err := s.bucket.PutObject(normalizedPath, reader, options...)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "上传文件到阿里云OSS失败: %s", normalizedPath)
		s.logger.Error("上传文件到阿里云OSS失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("文件上传到阿里云OSS成功", interfaces.String("path", normalizedPath))
	return nil
}

// Download 从阿里云OSS下载文件
//
// 描述: 从阿里云OSS存储下载文件数据
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	([]byte): 文件数据
//	(error): 下载错误信息
func (s *AliOSSStorage) Download(ctx context.Context, path string) ([]byte, error) {
	s.logger.Debug("正在从阿里云OSS下载文件", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("下载文件失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return nil, err
	}

	normalizedPath := s.normalizePath(path)

	// 获取对象
	object, err := s.bucket.GetObject(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "从阿里云OSS下载文件失败: %s", normalizedPath)
		s.logger.Error("从阿里云OSS下载文件失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return nil, wrappedErr
	}
	defer object.Close()

	// 读取数据
	data, err := io.ReadAll(object)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "读取阿里云OSS文件数据失败: %s", normalizedPath)
		s.logger.Error("读取阿里云OSS文件数据失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return nil, wrappedErr
	}

	s.logger.Info("从阿里云OSS下载文件成功", interfaces.String("path", normalizedPath), interfaces.Int("size", len(data)))
	return data, nil
}

// Delete 删除阿里云OSS文件
//
// 描述: 从阿里云OSS存储删除指定文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	error: 删除错误信息
func (s *AliOSSStorage) Delete(ctx context.Context, path string) error {
	s.logger.Debug("正在删除阿里云OSS文件", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("删除文件失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedPath := s.normalizePath(path)

	err := s.bucket.DeleteObject(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "删除阿里云OSS文件失败: %s", normalizedPath)
		s.logger.Error("删除阿里云OSS文件失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("阿里云OSS文件删除成功", interfaces.String("path", normalizedPath))
	return nil
}

// Exists 检查阿里云OSS文件是否存在
//
// 描述: 检查指定路径的文件是否在阿里云OSS中存在
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	(bool): 是否存在
//	(error): 检查错误信息
func (s *AliOSSStorage) Exists(ctx context.Context, path string) (bool, error) {
	s.logger.Debug("正在检查阿里云OSS文件是否存在", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("检查文件存在性失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return false, err
	}

	normalizedPath := s.normalizePath(path)

	exists, err := s.bucket.IsObjectExist(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "检查阿里云OSS文件存在性失败: %s", normalizedPath)
		s.logger.Error("检查阿里云OSS文件存在性失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return false, wrappedErr
	}

	s.logger.Debug("阿里云OSS文件存在性检查完成", interfaces.String("path", normalizedPath), interfaces.Bool("exists", exists))
	return exists, nil
}

// GetMetadata 获取阿里云OSS文件元数据
//
// 描述: 获取阿里云OSS文件的元数据信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//
// 返回:
//
//	(map[string]string): 文件元数据
//	(error): 获取错误信息
func (s *AliOSSStorage) GetMetadata(ctx context.Context, path string) (map[string]string, error) {
	s.logger.Debug("正在获取阿里云OSS文件元数据", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("获取元数据失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return nil, err
	}

	normalizedPath := s.normalizePath(path)

	// 获取对象元数据
	headers, err := s.bucket.GetObjectMeta(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "获取阿里云OSS文件元数据失败: %s", normalizedPath)
		s.logger.Error("获取阿里云OSS文件元数据失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return nil, wrappedErr
	}

	metadata := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			metadata[strings.ToLower(key)] = values[0]
		}
	}

	s.logger.Debug("获取阿里云OSS文件元数据成功", interfaces.String("path", normalizedPath))
	return metadata, nil
}

// SetMetadata 设置阿里云OSS文件元数据
//
// 描述: 设置阿里云OSS文件的元数据信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	metadata (map[string]string): 要设置的元数据
//
// 返回:
//
//	error: 设置错误信息
func (s *AliOSSStorage) SetMetadata(ctx context.Context, path string, metadata map[string]string) error {
	s.logger.Debug("正在设置阿里云OSS文件元数据", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("设置元数据失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedPath := s.normalizePath(path)

	// 阿里云OSS不支持直接修改元数据，需要复制文件来更新元数据
	// 先获取现有文件
	exists, err := s.Exists(ctx, path)
	if err != nil {
		return err // Exists方法内部已记录日志
	}
	if !exists {
		err := errors.Errorf("文件不存在: %s", normalizedPath)
		s.logger.Warn("设置元数据失败，文件不存在", interfaces.ErrorField(err), interfaces.String("path", normalizedPath))
		return err
	}

	// 准备复制选项
	var options []oss.Option
	for key, value := range metadata {
		if key == "content-type" {
			options = append(options, oss.ContentType(value))
		} else {
			options = append(options, oss.Meta(key, value))
		}
	}

	// 复制文件以更新元数据
	_, err = s.bucket.CopyObject(normalizedPath, normalizedPath, options...)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "设置阿里云OSS文件元数据失败: %s", normalizedPath)
		s.logger.Error("设置阿里云OSS文件元数据失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("阿里云OSS文件元数据设置成功", interfaces.String("path", normalizedPath))
	return nil
}

// List 列出阿里云OSS文件
//
// 描述: 列出指定前缀的阿里云OSS文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	prefix (string): 路径前缀
//
// 返回:
//
//	([]interfaces.RemoteFileInfo): 文件信息列表
//	(error): 列举错误信息
func (s *AliOSSStorage) List(ctx context.Context, prefix string) ([]interfaces.RemoteFileInfo, error) {
	s.logger.Debug("正在列出阿里云OSS文件", interfaces.String("prefix", prefix))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("列出文件失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return nil, err
	}

	normalizedPrefix := s.normalizePath(prefix)

	// 列举对象
	marker := ""
	var fileInfos []interfaces.RemoteFileInfo

	for {
		result, err := s.bucket.ListObjects(oss.Prefix(normalizedPrefix), oss.Marker(marker), oss.MaxKeys(1000))
		if err != nil {
			wrappedErr := errors.Wrapf(err, "列举阿里云OSS文件失败: %s", normalizedPrefix)
			s.logger.Error("列举阿里云OSS文件失败", interfaces.ErrorField(wrappedErr), interfaces.String("prefix", normalizedPrefix))
			return nil, wrappedErr
		}

		for _, object := range result.Objects {
			fileInfo := interfaces.RemoteFileInfo{
				Path:         object.Key,
				Size:         object.Size,
				ModTime:      object.LastModified,
				IsDir:        false,
				ContentType:  object.Type,
				ETag:         object.ETag,
				LastModified: object.LastModified,
				StorageClass: object.StorageClass,
				Owner:        object.Owner.DisplayName,
				ACL:          "", // 需要单独获取
			}
			fileInfos = append(fileInfos, fileInfo)
		}

		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}

	s.logger.Debug("列出阿里云OSS文件成功", interfaces.String("prefix", prefix), interfaces.Int("count", len(fileInfos)))
	return fileInfos, nil
}

// Copy 复制阿里云OSS文件
//
// 描述: 在阿里云OSS内部复制文件
//
// 参数:
//
//	ctx (context.Context): 上下文
//	srcPath (string): 源文件路径
//	destPath (string): 目标文件路径
//
// 返回:
//
//	error: 复制错误信息
func (s *AliOSSStorage) Copy(ctx context.Context, srcPath, destPath string) error {
	s.logger.Debug("正在复制阿里云OSS文件", interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("复制文件失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedSrcPath := s.normalizePath(srcPath)
	normalizedDestPath := s.normalizePath(destPath)

	_, err := s.bucket.CopyObject(normalizedSrcPath, normalizedDestPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "复制阿里云OSS文件失败: %s -> %s", normalizedSrcPath, normalizedDestPath)
		s.logger.Error("复制阿里云OSS文件失败", interfaces.ErrorField(wrappedErr), interfaces.String("source", normalizedSrcPath), interfaces.String("destination", normalizedDestPath))
		return wrappedErr
	}

	s.logger.Info("阿里云OSS文件复制成功", interfaces.String("source", normalizedSrcPath), interfaces.String("destination", normalizedDestPath))
	return nil
}

// Move 移动阿里云OSS文件
//
// 描述: 在阿里云OSS内部移动文件（复制后删除源文件）
//
// 参数:
//
//	ctx (context.Context): 上下文
//	srcPath (string): 源文件路径
//	destPath (string): 目标文件路径
//
// 返回:
//
//	error: 移动错误信息
func (s *AliOSSStorage) Move(ctx context.Context, srcPath, destPath string) error {
	s.logger.Debug("正在移动阿里云OSS文件", interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	// 先复制文件
	if err := s.Copy(ctx, srcPath, destPath); err != nil {
		return err // Copy方法内部已记录日志
	}

	// 删除源文件
	if err := s.Delete(ctx, srcPath); err != nil {
		// 移动操作中删除源文件失败是一个警告，因为文件已经复制成功
		s.logger.Warn("移动文件时删除源文件失败，但文件已成功复制", interfaces.ErrorField(err), interfaces.String("source", srcPath))
		// 不返回错误，因为主要操作（移动）已完成
	}

	s.logger.Info("阿里云OSS文件移动成功", interfaces.String("source", srcPath), interfaces.String("destination", destPath))
	return nil
}

// GetURL 获取阿里云OSS文件访问URL
//
// 描述: 获取阿里云OSS文件的访问URL，支持临时访问链接
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	expiration (time.Duration): 过期时间（0表示永久链接）
//
// 返回:
//
//	(string): 访问URL
//	(error): 获取错误信息
func (s *AliOSSStorage) GetURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	s.logger.Debug("正在获取阿里云OSS文件URL", interfaces.String("path", path), interfaces.String("expiration", expiration.String()))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("获取URL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return "", err
	}

	normalizedPath := s.normalizePath(path)

	if expiration > 0 {
		// 生成临时访问链接
		signedURL, err := s.bucket.SignURL(normalizedPath, oss.HTTPGet, int64(expiration.Seconds()))
		if err != nil {
			wrappedErr := errors.Wrapf(err, "生成阿里云OSS临时访问链接失败: %s", normalizedPath)
			s.logger.Error("生成阿里云OSS临时访问链接失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
			return "", wrappedErr
		}
		s.logger.Debug("生成阿里云OSS临时访问链接成功", interfaces.String("url", signedURL))
		return signedURL, nil
	} else {
		// 生成永久访问链接
		url := s.buildPublicURL(normalizedPath)
		s.logger.Debug("生成阿里云OSS永久访问链接成功", interfaces.String("url", url))
		return url, nil
	}
}

// GenerateFilePath 生成文件路径
//
// 描述: 为文件生成唯一的存储路径
//
// 参数:
//
//	originalName (string): 原始文件名
//
// 返回:
//
//	(string): 生成的文件路径
//	(error): 错误信息
func (s *AliOSSStorage) GenerateFilePath(originalName string) (string, error) {
	s.logger.Debug("正在生成文件路径", interfaces.String("original_name", originalName))
	ext := filepath.Ext(originalName)
	if ext == "" {
		return "", errors.New("文件没有扩展名")
	}

	// 生成随机文件名
	randomName, err := generateRandomString(32)
	if err != nil {
		wrappedErr := errors.Wrap(err, "生成随机文件名失败")
		s.logger.Error("生成随机文件名失败", interfaces.ErrorField(wrappedErr))
		return "", wrappedErr
	}

	// 生成日期路径
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	// 组合路径
	fileName := randomName + ext
	filePath := filepath.ToSlash(filepath.Join("files", datePath, fileName))

	s.logger.Debug("文件路径生成成功", interfaces.String("path", filePath))
	return filePath, nil
}

// normalizePath 标准化文件路径
//
// 描述: 标准化文件路径，添加基础路径前缀
//
// 参数:
//
//	path (string): 原始路径
//
// 返回:
//
//	string: 标准化后的路径
func (s *AliOSSStorage) normalizePath(path string) string {
	// 这个函数是内部辅助函数，日志可以在调用它的地方记录
	// 移除开头的斜杠
	path = strings.TrimPrefix(path, "/")

	// 添加基础路径前缀
	if s.config.BasePath != "" {
		basePath := strings.TrimSuffix(s.config.BasePath, "/")
		path = basePath + "/" + path
	}

	// 转换为正斜杠（OSS使用正斜杠）
	path = filepath.ToSlash(path)

	return path
}

// getContentType 根据文件扩展名获取内容类型
//
// 描述: 根据文件扩展名确定MIME类型
//
// 参数:
//
//	path (string): 文件路径
//
// 返回:
//
//	string: MIME类型
func (s *AliOSSStorage) getContentType(path string) string {
	// 这个函数是内部辅助函数，日志可以在调用它的地方记录
	ext := strings.ToLower(filepath.Ext(path))

	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}

// buildPublicURL 构建阿里云OSS公共访问URL
//
// 描述: 根据配置构建阿里云OSS文件的公共访问URL
//
// 参数:
//
//	path (string): 文件路径
//
// 返回:
//
//	string: 公共访问URL
func (s *AliOSSStorage) buildPublicURL(path string) string {
	// 这个函数是内部辅助函数，日志可以在调用它的地方记录
	if s.config.BucketDomain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(s.config.BucketDomain, "/"), strings.TrimLeft(path, "/"))
	}
	// 使用默认的阿里云OSS域名格式
	return fmt.Sprintf("https://%s.%s/%s", s.config.BucketName, s.config.Endpoint, strings.TrimLeft(path, "/"))
}

// generateRandomString 生成随机字符串
//
// 描述: 生成指定长度的随机十六进制字符串
//
// 参数:
//
//	length (int): 字符串长度
//
// 返回:
//
//	(string): 随机字符串
//	(error): 错误信息
func generateRandomString(length int) (string, error) {
	// 这个函数是内部辅助函数，日志可以在调用它的地方记录
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetPresignedUploadURL 获取预签名上传URL
//
// 描述: 生成用于直接上传的预签名URL
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	expiration (time.Duration): 过期时间
//
// 返回:
//
//	(string): 预签名上传URL
//	(error): 错误信息
func (s *AliOSSStorage) GetPresignedUploadURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	s.logger.Debug("正在生成预签名上传URL", interfaces.String("path", path), interfaces.String("expiration", expiration.String()))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("生成预签名上传URL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return "", err
	}

	normalizedPath := s.normalizePath(path)

	// 生成预签名上传URL
	signedURL, err := s.bucket.SignURL(normalizedPath, oss.HTTPPut, int64(expiration.Seconds()))
	if err != nil {
		wrappedErr := errors.Wrapf(err, "生成阿里云OSS预签名上传URL失败: %s", normalizedPath)
		s.logger.Error("生成阿里云OSS预签名上传URL失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return "", wrappedErr
	}

	s.logger.Info("生成预签名上传URL成功", interfaces.String("path", normalizedPath))
	return signedURL, nil
}

// SetFileACL 设置文件访问权限
//
// 描述: 设置阿里云OSS文件的访问权限
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	acl (string): 访问权限（private, public-read, public-read-write）
//
// 返回:
//
//	error: 设置错误信息
func (s *AliOSSStorage) SetFileACL(ctx context.Context, path string, acl string) error {
	s.logger.Debug("正在设置文件ACL", interfaces.String("path", path), interfaces.String("acl", acl))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("设置文件ACL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedPath := s.normalizePath(path)

	// 设置文件ACL
	var ossACL oss.ACLType
	switch acl {
	case "private":
		ossACL = oss.ACLPrivate
	case "public-read":
		ossACL = oss.ACLPublicRead
	case "public-read-write":
		ossACL = oss.ACLPublicReadWrite
	default:
		err := errors.Errorf("不支持的ACL类型: %s", acl)
		s.logger.Error("不支持的ACL类型", interfaces.ErrorField(err))
		return err
	}

	err := s.bucket.SetObjectACL(normalizedPath, ossACL)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "设置阿里云OSS文件ACL失败: %s", normalizedPath)
		s.logger.Error("设置阿里云OSS文件ACL失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("设置阿里云OSS文件ACL成功", interfaces.String("path", normalizedPath), interfaces.String("acl", acl))
	return nil
}

// BatchUpload 批量上传文件
//
// 描述: 批量上传多个文件到阿里云OSS
//
// 参数:
//
//	ctx (context.Context): 上下文
//	files (map[string][]byte): 文件路径和数据的映射
//	metadata (map[string]map[string]string): 每个文件的元数据
//
// 返回:
//
//	error: 批量上传错误信息
func (s *AliOSSStorage) BatchUpload(ctx context.Context, files map[string][]byte, metadata map[string]map[string]string) error {
	s.logger.Debug("正在批量上传文件", interfaces.Int("count", len(files)))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("批量上传失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	var uploadErrors []string

	for path, data := range files {
		fileMetadata := make(map[string]string)
		if metadata != nil && metadata[path] != nil {
			fileMetadata = metadata[path]
		}

		if err := s.Upload(ctx, path, data, fileMetadata); err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("文件 %s: %v", path, err))
			// 日志在Upload方法内部已经记录
		}
	}

	if len(uploadErrors) > 0 {
		err := errors.Errorf("批量上传部分失败: %s", strings.Join(uploadErrors, "; "))
		s.logger.Error("批量上传部分失败", interfaces.ErrorField(err))
		return err
	}

	s.logger.Info("批量上传成功", interfaces.Int("count", len(files)))
	return nil
}

// BatchDelete 批量删除文件
//
// 描述: 批量删除多个文件从阿里云OSS
//
// 参数:
//
//	ctx (context.Context): 上下文
//	paths ([]string): 要删除的文件路径列表
//
// 返回:
//
//	error: 批量删除错误信息
func (s *AliOSSStorage) BatchDelete(ctx context.Context, paths []string) error {
	s.logger.Debug("正在批量删除文件", interfaces.Int("count", len(paths)))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("批量删除失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	if len(paths) == 0 {
		s.logger.Debug("文件列表为空，无需删除")
		return nil
	}

	// 标准化路径
	normalizedPaths := make([]string, len(paths))
	for i, path := range paths {
		normalizedPaths[i] = s.normalizePath(path)
	}

	// 批量删除
	deleteResult, err := s.bucket.DeleteObjects(normalizedPaths)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "批量删除阿里云OSS文件失败")
		s.logger.Error("批量删除阿里云OSS文件失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	// 检查删除结果
	if len(deleteResult.DeletedObjects) != len(normalizedPaths) {
		s.logger.Warn("批量删除部分成功", interfaces.Int("requested", len(normalizedPaths)), interfaces.Int("deleted", len(deleteResult.DeletedObjects)))
	}

	s.logger.Info("批量删除成功", interfaces.Int("count", len(deleteResult.DeletedObjects)))
	return nil
}

// GetBucketInfo 获取存储桶信息
//
// 描述: 获取阿里云OSS存储桶的详细信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	(interfaces.RemoteBucketInfo): 存储桶信息
//	(error): 获取错误信息
func (s *AliOSSStorage) GetBucketInfo(ctx context.Context) (interfaces.RemoteBucketInfo, error) {
	s.logger.Debug("正在获取存储桶信息", interfaces.String("bucket_name", s.config.BucketName))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("获取存储桶信息失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return interfaces.RemoteBucketInfo{}, err
	}

	// 获取存储桶信息
	bucketInfo, err := s.client.GetBucketInfo(s.config.BucketName)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "获取阿里云OSS存储桶信息失败")
		s.logger.Error("获取阿里云OSS存储桶信息失败", interfaces.ErrorField(wrappedErr))
		return interfaces.RemoteBucketInfo{}, wrappedErr
	}

	// 统计文件数量和大小
	var totalSize int64
	var fileCount int64

	marker := ""
	for {
		result, err := s.bucket.ListObjects(oss.Marker(marker), oss.MaxKeys(1000))
		if err != nil {
			wrappedErr := errors.Wrapf(err, "列举阿里云OSS文件以统计信息失败")
			s.logger.Error("列举阿里云OSS文件以统计信息失败", interfaces.ErrorField(wrappedErr))
			return interfaces.RemoteBucketInfo{}, wrappedErr
		}

		for _, object := range result.Objects {
			totalSize += object.Size
			fileCount++
		}

		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
	}

	remoteBucketInfo := interfaces.RemoteBucketInfo{
		Name:         bucketInfo.BucketInfo.Name,
		Region:       bucketInfo.BucketInfo.Location,
		CreationDate: bucketInfo.BucketInfo.CreationDate,
		ACL:          string(bucketInfo.BucketInfo.ACL),
		StorageClass: string(bucketInfo.BucketInfo.StorageClass),
		Owner:        bucketInfo.BucketInfo.Owner.DisplayName,
		Size:         totalSize,
		ObjectCount:  fileCount,
		Metadata:     make(map[string]string), // OSS存储桶没有自定义元数据
	}

	s.logger.Info("获取存储桶信息成功", interfaces.String("name", remoteBucketInfo.Name), interfaces.Int64("object_count", remoteBucketInfo.ObjectCount))
	return remoteBucketInfo, nil
}

// GetPresignedDownloadURL 获取预签名下载URL
//
// 描述: 生成用于直接下载的预签名URL
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 文件路径
//	expiration (time.Duration): 过期时间
//
// 返回:
//
//	(string): 预签名下载URL
//	(error): 错误信息
func (s *AliOSSStorage) GetPresignedDownloadURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	s.logger.Debug("正在生成预签名下载URL", interfaces.String("path", path), interfaces.String("expiration", expiration.String()))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("生成预签名下载URL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return "", err
	}

	normalizedPath := s.normalizePath(path)

	// 生成预签名下载URL
	signedURL, err := s.bucket.SignURL(normalizedPath, oss.HTTPGet, int64(expiration.Seconds()))
	if err != nil {
		wrappedErr := errors.Wrapf(err, "生成阿里云OSS预签名下载URL失败: %s", normalizedPath)
		s.logger.Error("生成阿里云OSS预签名下载URL失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return "", wrappedErr
	}

	s.logger.Info("生成预签名下载URL成功", interfaces.String("path", normalizedPath))
	return signedURL, nil
}

// SetObjectACL 设置对象访问权限
//
// 描述: 设置阿里云OSS对象的访问权限
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 对象路径
//	acl (string): 访问权限
//
// 返回:
//
//	error: 设置错误信息
func (s *AliOSSStorage) SetObjectACL(ctx context.Context, path string, acl string) error {
	return s.SetFileACL(ctx, path, acl) // 复用已有的方法
}

// GetObjectACL 获取对象访问权限
//
// 描述: 获取阿里云OSS对象的访问权限
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 对象路径
//
// 返回:
//
//	(string): 访问权限
//	(error): 获取错误信息
func (s *AliOSSStorage) GetObjectACL(ctx context.Context, path string) (string, error) {
	s.logger.Debug("正在获取对象ACL", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("获取对象ACL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return "", err
	}

	normalizedPath := s.normalizePath(path)

	// 获取对象ACL
	result, err := s.bucket.GetObjectACL(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "获取阿里云OSS对象ACL失败: %s", normalizedPath)
		s.logger.Error("获取阿里云OSS对象ACL失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return "", wrappedErr
	}

	s.logger.Debug("获取对象ACL成功", interfaces.String("path", normalizedPath), interfaces.String("acl", string(result.ACL)))
	return string(result.ACL), nil
}

// ListBuckets 列出所有存储桶
//
// 描述: 列出阿里云OSS账户下的所有存储桶
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	([]interfaces.RemoteBucketInfo): 存储桶列表
//	(error): 列举错误信息
func (s *AliOSSStorage) ListBuckets(ctx context.Context) ([]interfaces.RemoteBucketInfo, error) {
	s.logger.Debug("正在列出所有存储桶")
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("列出存储桶失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return nil, err
	}

	// 列出所有存储桶
	result, err := s.client.ListBuckets()
	if err != nil {
		wrappedErr := errors.Wrapf(err, "列举阿里云OSS存储桶失败")
		s.logger.Error("列举阿里云OSS存储桶失败", interfaces.ErrorField(wrappedErr))
		return nil, wrappedErr
	}

	var buckets []interfaces.RemoteBucketInfo
	for _, bucket := range result.Buckets {
		bucketInfo := interfaces.RemoteBucketInfo{
			Name:         bucket.Name,
			Region:       bucket.Location,
			CreationDate: bucket.CreationDate,
			ACL:          "", // 需要单独获取
			StorageClass: string(bucket.StorageClass),
			Owner:        result.Owner.DisplayName,
			Metadata:     make(map[string]string),
		}
		buckets = append(buckets, bucketInfo)
	}

	s.logger.Info("列举存储桶成功", interfaces.Int("count", len(buckets)))
	return buckets, nil
}

// CreateBucket 创建存储桶
//
// 描述: 创建新的阿里云OSS存储桶
//
// 参数:
//
//	ctx (context.Context): 上下文
//	bucketName (string): 存储桶名称
//	config (map[string]interface{}): 存储桶配置
//
// 返回:
//
//	error: 创建错误信息
func (s *AliOSSStorage) CreateBucket(ctx context.Context, bucketName string, config map[string]interface{}) error {
	s.logger.Debug("正在创建存储桶", interfaces.String("bucket_name", bucketName))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("创建存储桶失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	// 准备创建选项
	var options []oss.Option

	// 设置ACL
	if acl, ok := config["acl"].(string); ok && acl != "" {
		switch acl {
		case "private":
			options = append(options, oss.ACL(oss.ACLPrivate))
		case "public-read":
			options = append(options, oss.ACL(oss.ACLPublicRead))
		case "public-read-write":
			options = append(options, oss.ACL(oss.ACLPublicReadWrite))
		}
	}

	// 设置存储类型
	if storageClass, ok := config["storage_class"].(string); ok && storageClass != "" {
		switch storageClass {
		case "Standard":
			options = append(options, oss.StorageClass(oss.StorageStandard))
		case "IA":
			options = append(options, oss.StorageClass(oss.StorageIA))
		case "Archive":
			options = append(options, oss.StorageClass(oss.StorageArchive))
		case "ColdArchive":
			options = append(options, oss.StorageClass(oss.StorageColdArchive))
		}
	}

	// 创建存储桶
	err := s.client.CreateBucket(bucketName, options...)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "创建阿里云OSS存储桶失败: %s", bucketName)
		s.logger.Error("创建阿里云OSS存储桶失败", interfaces.ErrorField(wrappedErr), interfaces.String("bucket_name", bucketName))
		return wrappedErr
	}

	s.logger.Info("创建存储桶成功", interfaces.String("bucket_name", bucketName))
	return nil
}

// DeleteBucket 删除存储桶
//
// 描述: 删除阿里云OSS存储桶
//
// 参数:
//
//	ctx (context.Context): 上下文
//	bucketName (string): 存储桶名称
//
// 返回:
//
//	error: 删除错误信息
func (s *AliOSSStorage) DeleteBucket(ctx context.Context, bucketName string) error {
	s.logger.Debug("正在删除存储桶", interfaces.String("bucket_name", bucketName))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("删除存储桶失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	// 删除存储桶
	err := s.client.DeleteBucket(bucketName)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "删除阿里云OSS存储桶失败: %s", bucketName)
		s.logger.Error("删除阿里云OSS存储桶失败", interfaces.ErrorField(wrappedErr), interfaces.String("bucket_name", bucketName))
		return wrappedErr
	}

	s.logger.Info("删除存储桶成功", interfaces.String("bucket_name", bucketName))
	return nil
}

// SetBucketACL 设置存储桶访问权限
//
// 描述: 设置阿里云OSS存储桶的访问权限
//
// 参数:
//
//	ctx (context.Context): 上下文
//	bucketName (string): 存储桶名称
//	acl (string): 访问权限
//
// 返回:
//
//	error: 设置错误信息
func (s *AliOSSStorage) SetBucketACL(ctx context.Context, bucketName string, acl string) error {
	s.logger.Debug("正在设置存储桶ACL", interfaces.String("bucket_name", bucketName), interfaces.String("acl", acl))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("设置存储桶ACL失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	// 设置存储桶ACL
	var ossACL oss.ACLType
	switch acl {
	case "private":
		ossACL = oss.ACLPrivate
	case "public-read":
		ossACL = oss.ACLPublicRead
	case "public-read-write":
		ossACL = oss.ACLPublicReadWrite
	default:
		err := errors.Errorf("不支持的ACL类型: %s", acl)
		s.logger.Error("不支持的ACL类型", interfaces.ErrorField(err))
		return err
	}

	err := s.client.SetBucketACL(bucketName, ossACL)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "设置阿里云OSS存储桶ACL失败: %s", bucketName)
		s.logger.Error("设置阿里云OSS存储桶ACL失败", interfaces.ErrorField(wrappedErr), interfaces.String("bucket_name", bucketName))
		return wrappedErr
	}

	s.logger.Info("设置存储桶ACL成功", interfaces.String("bucket_name", bucketName), interfaces.String("acl", acl))
	return nil
}

// GetObjectInfo 获取对象详细信息
//
// 描述: 获取阿里云OSS对象的详细信息
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 对象路径
//
// 返回:
//
//	(interfaces.RemoteObjectInfo): 对象信息
//	(error): 获取错误信息
func (s *AliOSSStorage) GetObjectInfo(ctx context.Context, path string) (interfaces.RemoteObjectInfo, error) {
	s.logger.Debug("正在获取对象详细信息", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("获取对象详细信息失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return interfaces.RemoteObjectInfo{}, err
	}

	normalizedPath := s.normalizePath(path)

	// 获取对象元数据
	metadata, err := s.GetMetadata(ctx, path)
	if err != nil {
		return interfaces.RemoteObjectInfo{}, err
	}

	// 获取对象ACL
	acl, err := s.GetObjectACL(ctx, path)
	if err != nil {
		acl = "" // 如果获取失败，设为空字符串
	}

	objectInfo := interfaces.RemoteObjectInfo{
		Path:         normalizedPath,
		Size:         0, // 需要从metadata中解析
		ContentType:  metadata["content-type"],
		ETag:         metadata["etag"],
		StorageClass: metadata["x-oss-storage-class"],
		ACL:          acl,
		Owner:        metadata["x-oss-owner"],
		Metadata:     metadata,
		IsSymlink:    false,
		VersionId:    metadata["x-oss-version-id"],
	}

	// 解析大小
	if sizeStr, ok := metadata["content-length"]; ok {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			objectInfo.Size = size
		}
	}

	// 解析最后修改时间
	if lastModStr, ok := metadata["last-modified"]; ok {
		if lastMod, err := time.Parse(time.RFC1123, lastModStr); err == nil {
			objectInfo.LastModified = lastMod
		}
	}

	s.logger.Debug("获取对象详细信息成功", interfaces.String("path", normalizedPath))
	return objectInfo, nil
}

// RestoreObject 恢复归档对象
//
// 描述: 恢复阿里云OSS归档类型的对象
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 对象路径
//	days (int): 恢复天数
//
// 返回:
//
//	error: 恢复错误信息
func (s *AliOSSStorage) RestoreObject(ctx context.Context, path string, days int) error {
	s.logger.Debug("正在恢复归档对象", interfaces.String("path", path), interfaces.Int("days", days))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("恢复归档对象失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedPath := s.normalizePath(path)

	// 恢复归档对象
	err := s.bucket.RestoreObject(normalizedPath)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "恢复阿里云OSS归档对象失败: %s", normalizedPath)
		s.logger.Error("恢复阿里云OSS归档对象失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("恢复归档对象成功", interfaces.String("path", normalizedPath), interfaces.Int("days", days))
	return nil
}

// SetObjectStorageClass 设置对象存储类型
//
// 描述: 设置阿里云OSS对象的存储类型
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): 对象路径
//	storageClass (string): 存储类型
//
// 返回:
//
//	error: 设置错误信息
func (s *AliOSSStorage) SetObjectStorageClass(ctx context.Context, path string, storageClass string) error {
	s.logger.Debug("正在设置对象存储类型", interfaces.String("path", path), interfaces.String("storage_class", storageClass))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("设置对象存储类型失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return err
	}

	normalizedPath := s.normalizePath(path)

	// 设置对象存储类型
	var ossStorageClass oss.StorageClassType
	switch storageClass {
	case "Standard":
		ossStorageClass = oss.StorageStandard
	case "IA":
		ossStorageClass = oss.StorageIA
	case "Archive":
		ossStorageClass = oss.StorageArchive
	case "ColdArchive":
		ossStorageClass = oss.StorageColdArchive
	default:
		err := errors.Errorf("不支持的存储类型: %s", storageClass)
		s.logger.Error("不支持的存储类型", interfaces.ErrorField(err))
		return err
	}

	// 通过复制对象来修改存储类型
	_, err := s.bucket.CopyObject(normalizedPath, normalizedPath, oss.ObjectStorageClass(ossStorageClass))
	if err != nil {
		wrappedErr := errors.Wrapf(err, "设置阿里云OSS对象存储类型失败: %s", normalizedPath)
		s.logger.Error("设置阿里云OSS对象存储类型失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return wrappedErr
	}

	s.logger.Info("设置对象存储类型成功", interfaces.String("path", normalizedPath), interfaces.String("storage_class", storageClass))
	return nil
}

// GetConfig returns the current configuration for the AliOSSStorage.
func (s *AliOSSStorage) GetConfig() *AliOSSConfig {
	return s.config
}

// Driver returns the driver name.
func (s *AliOSSStorage) Driver() string {
	return "oss"
}

// Get 从阿里云OSS下载文件到本地临时文件
//
// 描述: 从阿里云OSS下载文件并保存到本地临时文件，返回文件句柄
//
// 参数:
//
//	ctx (context.Context): 上下文
//	path (string): OSS文件路径
//
// 返回:
//
//	(*os.File): 本地临时文件句柄
//	(error): 错误信息
func (s *AliOSSStorage) Get(ctx context.Context, path string) (*os.File, error) {
	s.logger.Debug("正在从阿里云OSS下载文件到本地临时文件", interfaces.String("path", path))
	if !s.connected {
		err := errors.New("阿里云OSS未连接")
		s.logger.Error("下载文件到本地失败，阿里云OSS未连接", interfaces.ErrorField(err))
		return nil, err
	}

	normalizedPath := s.normalizePath(path)

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "oss-download-*")
	if err != nil {
		wrappedErr := errors.Wrap(err, "创建临时文件失败")
		s.logger.Error("创建临时文件失败", interfaces.ErrorField(wrappedErr))
		return nil, wrappedErr
	}

	// 使用 GetObjectToFile 下载文件到临时文件
	err = s.bucket.GetObjectToFile(normalizedPath, tempFile.Name())
	if err != nil {
		// 下载失败时清理临时文件
		tempFile.Close()
		os.Remove(tempFile.Name())
		wrappedErr := errors.Wrapf(err, "从阿里云OSS下载文件到本地失败: %s", normalizedPath)
		s.logger.Error("从阿里云OSS下载文件到本地失败", interfaces.ErrorField(wrappedErr), interfaces.String("path", normalizedPath))
		return nil, wrappedErr
	}

	// 重新打开文件用于读取
	tempFile.Close()
	downloadedFile, err := os.Open(tempFile.Name())
	if err != nil {
		// 打开失败时清理临时文件
		os.Remove(tempFile.Name())
		wrappedErr := errors.Wrap(err, "打开下载的临时文件失败")
		s.logger.Error("打开下载的临时文件失败", interfaces.ErrorField(wrappedErr))
		return nil, wrappedErr
	}

	s.logger.Info("从阿里云OSS下载文件到本地成功", interfaces.String("path", normalizedPath), interfaces.String("temp_file", downloadedFile.Name()))
	return downloadedFile, nil
}
