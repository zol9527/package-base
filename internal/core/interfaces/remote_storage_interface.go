package interfaces

import (
	"context"
	"time"
)

// RemoteBucketInfo 远程存储桶信息
//
// 描述: 表示远程存储桶的详细信息
type RemoteBucketInfo struct {
	Name         string            `json:"name"`          // 存储桶名称
	Region       string            `json:"region"`        // 区域
	CreationDate time.Time         `json:"creation_date"` // 创建时间
	ACL          string            `json:"acl"`           // 访问权限
	StorageClass string            `json:"storage_class"` // 存储类型
	Owner        string            `json:"owner"`         // 所有者
	Metadata     map[string]string `json:"metadata"`      // 元数据
	Size         int64             `json:"size"`          // 大小（字节）
	ObjectCount  int64             `json:"object_count"`  // 对象数量
}

// RemoteObjectInfo 远程对象信息
//
// 描述: 表示远程存储中对象的详细信息
type RemoteObjectInfo struct {
	Path         string            `json:"path"`          // 对象路径
	Size         int64             `json:"size"`          // 对象大小（字节）
	ContentType  string            `json:"content_type"`  // 内容类型
	ETag         string            `json:"etag"`          // ETag
	LastModified time.Time         `json:"last_modified"` // 最后修改时间
	StorageClass string            `json:"storage_class"` // 存储类型
	ACL          string            `json:"acl"`           // 访问权限
	Owner        string            `json:"owner"`         // 所有者
	Metadata     map[string]string `json:"metadata"`      // 元数据
	IsSymlink    bool              `json:"is_symlink"`    // 是否为符号链接
	VersionId    string            `json:"version_id"`    // 版本ID
}

// RemoteFileInfo 远程文件信息
//
// 描述: 表示远程存储中文件的详细信息
type RemoteFileInfo struct {
	Path         string            `json:"path"`          // 文件路径
	Size         int64             `json:"size"`          // 文件大小（字节）
	ModTime      time.Time         `json:"mod_time"`      // 修改时间
	IsDir        bool              `json:"is_dir"`        // 是否为目录
	ContentType  string            `json:"content_type"`  // 内容类型
	Metadata     map[string]string `json:"metadata"`      // 元数据
	ETag         string            `json:"etag"`          // ETag
	LastModified time.Time         `json:"last_modified"` // 最后修改时间
	StorageClass string            `json:"storage_class"` // 存储类型
	Owner        string            `json:"owner"`         // 所有者
	ACL          string            `json:"acl"`           // 访问权限
}

// RemoteStorage 远程存储接口
//
// 描述: 定义远程存储操作，专门用于云存储服务（如阿里云OSS、AWS S3等）
type RemoteOssStorage interface {
	BaseStorage

	// Upload 上传文件到远程存储
	Upload(ctx context.Context, path string, data []byte, metadata map[string]string) error

	// Download 从远程存储下载文件
	Download(ctx context.Context, path string) ([]byte, error)

	// Delete 删除远程存储中的文件
	Delete(ctx context.Context, path string) error

	// Exists 检查远程存储中文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetMetadata 获取远程文件元数据
	GetMetadata(ctx context.Context, path string) (map[string]string, error)

	// SetMetadata 设置远程文件元数据
	SetMetadata(ctx context.Context, path string, metadata map[string]string) error

	// List 列出远程存储中的文件
	List(ctx context.Context, prefix string) ([]RemoteFileInfo, error)

	// Copy 在远程存储内复制文件
	Copy(ctx context.Context, srcPath, destPath string) error

	// Move 在远程存储内移动文件
	Move(ctx context.Context, srcPath, destPath string) error

	// GetURL 获取文件访问URL（支持临时链接和永久链接）
	GetURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// GenerateFilePath 生成唯一的文件存储路径
	GenerateFilePath(originalName string) (string, error)

	// GetPresignedUploadURL 获取预签名上传URL
	GetPresignedUploadURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// GetPresignedDownloadURL 获取预签名下载URL
	GetPresignedDownloadURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// SetObjectACL 设置对象访问权限
	SetObjectACL(ctx context.Context, path string, acl string) error

	// GetObjectACL 获取对象访问权限
	GetObjectACL(ctx context.Context, path string) (string, error)

	// BatchUpload 批量上传文件
	BatchUpload(ctx context.Context, files map[string][]byte, metadata map[string]map[string]string) error

	// BatchDelete 批量删除文件
	BatchDelete(ctx context.Context, paths []string) error

	// GetBucketInfo 获取存储桶信息
	GetBucketInfo(ctx context.Context) (RemoteBucketInfo, error)

	// ListBuckets 列出所有存储桶
	ListBuckets(ctx context.Context) ([]RemoteBucketInfo, error)

	// CreateBucket 创建存储桶
	CreateBucket(ctx context.Context, bucketName string, config map[string]interface{}) error

	// DeleteBucket 删除存储桶
	DeleteBucket(ctx context.Context, bucketName string) error

	// SetBucketACL 设置存储桶访问权限
	SetBucketACL(ctx context.Context, bucketName string, acl string) error

	// GetObjectInfo 获取对象详细信息
	GetObjectInfo(ctx context.Context, path string) (RemoteObjectInfo, error)

	// RestoreObject 恢复归档对象
	RestoreObject(ctx context.Context, path string, days int) error

	// SetObjectStorageClass 设置对象存储类型
	SetObjectStorageClass(ctx context.Context, path string, storageClass string) error
}
