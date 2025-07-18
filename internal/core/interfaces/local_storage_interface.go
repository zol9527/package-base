package interfaces

import (
	"context"
	"mime/multipart"
	"os"
	"time"

	"gorm.io/gorm"
)

// DocumentStorage 文档存储接口
//
// 描述: 定义文档存储操作，适用于 JSON 文档、配置文件等场景
type DocumentStorage interface {
	BaseStorage

	// Store 存储文档
	Store(ctx context.Context, collection, id string, document interface{}) error

	// Retrieve 检索文档
	Retrieve(ctx context.Context, collection, id string, result interface{}) error

	// Update 更新文档
	Update(ctx context.Context, collection, id string, document interface{}) error

	// Delete 删除文档
	Delete(ctx context.Context, collection, id string) error

	// Exists 检查文档是否存在
	Exists(ctx context.Context, collection, id string) (bool, error)

	// List 列出文档
	List(ctx context.Context, collection string, filter map[string]interface{}) ([]interface{}, error)

	// Query 查询文档
	Query(ctx context.Context, collection string, query map[string]interface{}) ([]interface{}, error)

	// CreateCollection 创建集合
	CreateCollection(ctx context.Context, collection string) error

	// DropCollection 删除集合
	DropCollection(ctx context.Context, collection string) error
}

// FileInfo 文件信息
//
// 描述: 表示文件的基本信息
type FileInfo struct {
	Path         string            `json:"path"`          // 文件路径
	Size         int64             `json:"size"`          // 文件大小（字节）
	ModTime      time.Time         `json:"mod_time"`      // 修改时间
	IsDir        bool              `json:"is_dir"`        // 是否为目录
	ContentType  string            `json:"content_type"`  // 内容类型
	Metadata     map[string]string `json:"metadata"`      // 元数据
	ETag         string            `json:"etag"`          // ETag
	LastModified time.Time         `json:"last_modified"` // 最后修改时间
}

// FileStorage 文件存储接口
//
// 描述: 定义文件存储操作，适用于文件上传、静态资源存储等场景
type FileStorage interface {
	BaseStorage

	// Upload 上传文件
	Upload(ctx context.Context, path string, data []byte, metadata map[string]string) error

	// Download 下载文件
	Download(ctx context.Context, path string) ([]byte, error)

	// Delete 删除文件
	Delete(ctx context.Context, path string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, path string) (bool, error)

	// GetMetadata 获取文件元数据
	GetMetadata(ctx context.Context, path string) (map[string]string, error)

	// SetMetadata 设置文件元数据
	SetMetadata(ctx context.Context, path string, metadata map[string]string) error

	// List 列出文件
	List(ctx context.Context, prefix string) ([]FileInfo, error)

	// Copy 复制文件
	Copy(ctx context.Context, srcPath, destPath string) error

	// Move 移动文件
	Move(ctx context.Context, srcPath, destPath string) error

	// GetURL 获取文件访问URL
	GetURL(ctx context.Context, path string, expiration time.Duration) (string, error)

	// Save 保存文件
	Save(file *multipart.FileHeader) (string, error)

	// Get 获取文件
	Get(ctx context.Context, path string) (*os.File, error)

	// Driver 返回存储驱动名称
	Driver() string
}

// RelationalStorage 关系型存储接口
//
// 描述: 定义关系型数据库存储操作，基于 GORM 的抽象
type RelationalStorage interface {
	BaseStorage

	// GetDB 获取 gorm.DB 实例
	GetDB() *gorm.DB

	// Create 创建记录
	Create(ctx context.Context, value interface{}) error

	// FindByID 根据ID查找记录
	FindByID(ctx context.Context, id interface{}, result interface{}) error

	// Find 查找记录
	Find(ctx context.Context, result interface{}, conditions ...interface{}) error

	// Update 更新记录
	Update(ctx context.Context, value interface{}) error

	// Delete 删除记录
	Delete(ctx context.Context, value interface{}) error

	// Count 统计记录数量
	Count(ctx context.Context, model interface{}, conditions ...interface{}) (int64, error)

	// Exists 检查记录是否存在
	Exists(ctx context.Context, model interface{}, conditions ...interface{}) (bool, error)

	// Transaction 执行事务
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error

	// Raw 执行原生SQL
	Raw(ctx context.Context, sql string, values ...interface{}) error

	// Migrate 迁移表结构
	Migrate(ctx context.Context, models ...interface{}) error

	// InitializeModels 初始化所有模型
	InitializeModels(ctx context.Context) error

	// MigrateSpecificModels 迁移指定模型
	MigrateSpecificModels(ctx context.Context, models ...interface{}) error

	// CheckTableExists 检查表是否存在
	CheckTableExists(ctx context.Context, model interface{}) (bool, error)

	// DropTables 删除指定表
	DropTables(ctx context.Context, models ...interface{}) error
}

// CacheStorage 缓存存储接口
//
// 描述: 定义缓存存储操作，提供分层缓存功能
type CacheStorage interface {
	KeyValueStorage

	// SetWithTags 设置带标签的缓存
	SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags []string) error

	// DeleteByTags 根据标签删除缓存
	DeleteByTags(ctx context.Context, tags []string) error

	// GetMulti 批量获取
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)

	// SetMulti 批量设置
	SetMulti(ctx context.Context, items map[string]interface{}, expiration time.Duration) error

	// DeleteMulti 批量删除
	DeleteMulti(ctx context.Context, keys []string) error

	// Increment 递增
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement 递减
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Lock 分布式锁
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)

	// Unlock 释放分布式锁
	Unlock(ctx context.Context, key string) error
}
