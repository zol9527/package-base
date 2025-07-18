package storages

import (
	"context"
	"fmt"
	"net/url"
	"package-base/internal/core/interfaces"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// MySQLConfig MySQL配置结构
//
// 描述: MySQL数据库连接配置
type MySQLConfig struct {
	Host            string        `json:"host"`               // 主机地址
	Port            int           `json:"port"`               // 端口号
	Username        string        `json:"username"`           // 用户名
	Password        string        `json:"password"`           // 密码
	Database        string        `json:"database"`           // 数据库名
	Charset         string        `json:"charset"`            // 字符集
	ParseTime       bool          `json:"parse_time"`         // 是否解析时间
	Loc             string        `json:"loc"`                // 时区
	MaxOpenConns    int           `json:"max_open_conns"`     // 最大开放连接数
	MaxIdleConns    int           `json:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`  // 连接最大生存时间
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"` // 连接最大空闲时间
	LogLevel        string        `json:"log_level"`          // 日志级别
}

// GetType 获取存储类型
//
// 描述: 返回存储类型标识
//
// 返回:
//
//	string: 存储类型
func (c *MySQLConfig) GetType() string {
	return "mysql"
}

// Validate 验证配置
//
// 描述: 验证MySQL配置的有效性
//
// 返回:
//
//	error: 验证错误信息
func (c *MySQLConfig) Validate() error {
	if c.Host == "" {
		return errors.New("MySQL主机地址不能为空")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("MySQL端口号必须在1-65535之间")
	}
	if c.Username == "" {
		return errors.New("MySQL用户名不能为空")
	}
	if c.Database == "" {
		return errors.New("MySQL数据库名不能为空")
	}

	// 设置默认值
	if c.Charset == "" {
		c.Charset = "utf8mb4"
	}
	if c.Loc == "" {
		c.Loc = "Local"
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 100
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 10
	}
	if c.ConnMaxLifetime <= 0 {
		c.ConnMaxLifetime = time.Hour
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = time.Minute * 10
	}

	return nil
}

// GetConnectionString 获取连接字符串
//
// 描述: 构建MySQL连接字符串
//
// 返回:
//
//	string: 连接字符串
func (c *MySQLConfig) GetConnectionString() string {
	// 对时区参数进行URL编码
	loc := url.QueryEscape(c.Loc)

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
		c.ParseTime,
		loc,
	)
}

// MySQLStorage MySQL存储实现
//
// 描述: 基于GORM的MySQL关系型存储实现
type MySQLStorage struct {
	db     *gorm.DB
	config *MySQLConfig
	logger interfaces.Logger
}

// NewMySQLStorage 创建MySQL存储实例
//
// 描述: 根据配置创建MySQL存储实例
//
// 参数:
//
//	config (*MySQLConfig): MySQL配置
//	loggerInstance (interfaces.Logger): 日志记录器
//
// 返回:
//
//	storageinterfaces.RelationalStorage: 关系型存储接口
//	error: 错误信息
func NewMySQLStorage(config *MySQLConfig, loggerInstance interfaces.Logger) (interfaces.RelationalStorage, error) {
	if config == nil {
		return nil, errors.New("MySQL配置不能为空")
	}
	if loggerInstance == nil {
		return nil, errors.New("日志记录器不能为空")
	}

	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "MySQL配置验证失败")
	}

	storage := &MySQLStorage{
		config: config,
		logger: loggerInstance.With(interfaces.String("component", "mysql-storage")),
	}

	// 创建连接
	if err := storage.Connect(context.Background()); err != nil {
		return nil, errors.Wrap(err, "创建MySQL连接失败")
	}

	storage.logger.Info("MySQL存储已初始化", interfaces.String("host", config.Host), interfaces.Int("port", config.Port), interfaces.String("database", config.Database))
	return storage, nil
}

// Connect 连接到MySQL
//
// 描述: 建立到MySQL数据库的连接
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Connect(ctx context.Context) error {
	s.logger.Debug("正在连接到MySQL数据库...")
	if s.db != nil {
		s.logger.Debug("  - 连接已存在，跳过连接过程")
		return nil
	}

	// 设置日志级别
	var logLevel gormlogger.LogLevel
	switch s.config.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Error
	}

	// 创建GORM配置
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	}

	// 创建数据库连接
	s.logger.Debug("  - 正在打开GORM连接...")
	db, err := gorm.Open(mysql.Open(s.config.GetConnectionString()), gormConfig)
	if err != nil {
		wrappedErr := errors.Wrap(err, "创建GORM连接失败")
		s.logger.Error("创建GORM连接失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	// 获取底层sql.DB实例并配置连接池
	s.logger.Debug("  - 正在获取底层sql.DB实例并配置连接池...")
	sqlDB, err := db.DB()
	if err != nil {
		wrappedErr := errors.Wrap(err, "获取sql.DB实例失败")
		s.logger.Error("获取sql.DB实例失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(s.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(s.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(s.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)

	// 测试连接
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.logger.Debug("  - 正在Ping数据库以验证连接...")
	if err := sqlDB.PingContext(pingCtx); err != nil {
		wrappedErr := errors.Wrap(err, "MySQL连接测试失败")
		s.logger.Error("MySQL连接测试失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	s.db = db
	s.logger.Debug("MySQL连接成功")
	return nil
}

// Close 关闭连接
//
// 描述: 关闭MySQL数据库连接
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Close() error {
	s.logger.Debug("正在关闭MySQL连接...")
	if s.db == nil {
		s.logger.Debug("  - 连接不存在，无需关闭")
		return nil
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		wrappedErr := errors.Wrap(err, "获取sql.DB实例失败")
		s.logger.Error("关闭MySQL连接时获取sql.DB实例失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	if err := sqlDB.Close(); err != nil {
		wrappedErr := errors.Wrap(err, "关闭MySQL连接失败")
		s.logger.Error("关闭MySQL连接失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	s.db = nil
	s.logger.Debug("MySQL连接已关闭")
	return nil
}

// Ping 测试连接状态
//
// 描述: 测试MySQL连接是否正常
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Ping(ctx context.Context) error {
	s.logger.Debug("Pinging MySQL数据库...")
	if s.db == nil {
		err := errors.New("MySQL连接未建立")
		s.logger.Error("Ping失败，连接未建立", interfaces.ErrorField(err))
		return err
	}

	sqlDB, err := s.db.DB()
	if err != nil {
		wrappedErr := errors.Wrap(err, "获取sql.DB实例失败")
		s.logger.Error("Ping时获取sql.DB实例失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		wrappedErr := errors.Wrap(err, "MySQL ping失败")
		s.logger.Error("Ping MySQL数据库失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}

	s.logger.Debug("Ping MySQL数据库成功")
	return nil
}

// IsHealthy 检查存储健康状态
//
// 描述: 检查MySQL存储是否健康
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	bool: 是否健康
func (s *MySQLStorage) IsHealthy(ctx context.Context) bool {
	s.logger.Debug("正在检查MySQL健康状态...")
	isHealthy := s.Ping(ctx) == nil
	s.logger.Debug("MySQL健康状态检查完成", interfaces.Bool("is_healthy", isHealthy))
	return isHealthy
}

// Create 创建记录
//
// 描述: 创建新的数据库记录
//
// 参数:
//
//	ctx (context.Context): 上下文
//	value (interface{}): 要创建的记录
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Create(ctx context.Context, value interface{}) error {
	s.logger.Debug("正在创建记录...", interfaces.Any("value", value))
	if s.db == nil {
		err := errors.New("MySQL连接未建立")
		s.logger.Error("创建记录失败，MySQL连接未建立", interfaces.ErrorField(err))
		return err
	}

	if value == nil {
		return errors.New("创建的记录不能为空")
	}

	result := s.db.WithContext(ctx).Create(value)
	if result.Error != nil {
		wrappedErr := errors.Wrap(result.Error, "创建记录失败")
		s.logger.Error("创建记录失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", value)))
		return wrappedErr
	}

	s.logger.Debug("创建记录成功", interfaces.Int64("rows_affected", result.RowsAffected))
	return nil
}

// FindByID 根据ID查找记录
//
// 描述: 根据主键ID查找单条记录
//
// 参数:
//
//	ctx (context.Context): 上下文
//	id (interface{}): 主键ID
//	result (interface{}): 结果接收对象
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) FindByID(ctx context.Context, id interface{}, result interface{}) error {
	s.logger.Debug("正在通过ID查找记录", interfaces.String("model", fmt.Sprintf("%T", result)), interfaces.Any("id", id))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if id == nil {
		return errors.New("查找的ID不能为空")
	}

	if result == nil {
		return errors.New("结果接收对象不能为空")
	}

	err := s.db.WithContext(ctx).First(result, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wrappedErr := errors.Wrap(err, "记录不存在")
			s.logger.Warn("通过ID查找记录失败：记录不存在", interfaces.String("model", fmt.Sprintf("%T", result)), interfaces.Any("id", id))
			return wrappedErr
		}
		wrappedErr := errors.Wrap(err, "查找记录失败")
		s.logger.Error("通过ID查找记录失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", result)), interfaces.Any("id", id))
		return wrappedErr
	}

	s.logger.Debug("通过ID查找记录成功", interfaces.String("model", fmt.Sprintf("%T", result)), interfaces.Any("id", id))
	return nil
}

// Find 查找记录
//
// 描述: 根据条件查找记录
//
// 参数:
//
//	ctx (context.Context): 上下文
//	result (interface{}): 结果接收对象
//	conditions (...interface{}): 查找条件
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Find(ctx context.Context, result interface{}, conditions ...interface{}) error {
	s.logger.Debug("正在查找记录", interfaces.String("model", fmt.Sprintf("%T", result)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if result == nil {
		return errors.New("结果接收对象不能为空")
	}

	query := s.db.WithContext(ctx)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	err := query.Find(result).Error
	if err != nil {
		wrappedErr := errors.Wrap(err, "查找记录失败")
		s.logger.Error("查找记录失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", result)))
		return wrappedErr
	}

	s.logger.Debug("查找记录成功", interfaces.String("model", fmt.Sprintf("%T", result)))
	return nil
}

// Update 更新记录
//
// 描述: 更新数据库记录
//
// 参数:
//
//	ctx (context.Context): 上下文
//	value (interface{}): 要更新的记录
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Update(ctx context.Context, value interface{}) error {
	s.logger.Debug("正在更新记录", interfaces.String("model", fmt.Sprintf("%T", value)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if value == nil {
		return errors.New("更新的记录不能为空")
	}

	result := s.db.WithContext(ctx).Save(value)
	if result.Error != nil {
		wrappedErr := errors.Wrap(result.Error, "更新记录失败")
		s.logger.Error("更新记录失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", value)))
		return wrappedErr
	}

	s.logger.Debug("更新记录成功", interfaces.Int64("rows_affected", result.RowsAffected))
	return nil
}

// Delete 删除记录
//
// 描述: 删除数据库记录
//
// 参数:
//
//	ctx (context.Context): 上下文
//	value (interface{}): 要删除的记录
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Delete(ctx context.Context, value interface{}) error {
	s.logger.Debug("正在删除记录", interfaces.String("model", fmt.Sprintf("%T", value)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if value == nil {
		return errors.New("删除的记录不能为空")
	}

	result := s.db.WithContext(ctx).Delete(value)
	if result.Error != nil {
		wrappedErr := errors.Wrap(result.Error, "删除记录失败")
		s.logger.Error("删除记录失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", value)))
		return wrappedErr
	}

	s.logger.Debug("删除记录成功", interfaces.Int64("rows_affected", result.RowsAffected))
	return nil
}

// Count 统计记录数量
//
// 描述: 统计符合条件的记录数量
//
// 参数:
//
//	ctx (context.Context): 上下文
//	model (interface{}): 模型对象
//	conditions (...interface{}): 查询条件
//
// 返回:
//
//	int64: 记录数量
//	error: 错误信息
func (s *MySQLStorage) Count(ctx context.Context, model interface{}, conditions ...interface{}) (int64, error) {
	s.logger.Debug("正在统计记录数量", interfaces.String("model", fmt.Sprintf("%T", model)))
	if s.db == nil {
		return 0, errors.New("MySQL连接未建立")
	}

	if model == nil {
		return 0, errors.New("模型对象不能为空")
	}

	query := s.db.WithContext(ctx).Model(model)
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		wrappedErr := errors.Wrap(err, "统计记录数量失败")
		s.logger.Error("统计记录数量失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", fmt.Sprintf("%T", model)))
		return 0, wrappedErr
	}

	s.logger.Debug("统计记录数量成功", interfaces.String("model", fmt.Sprintf("%T", model)), interfaces.Int64("count", count))
	return count, nil
}

// Exists 检查记录是否存在
//
// 描述: 检查符合条件的记录是否存在
//
// 参数:
//
//	ctx (context.Context): 上下文
//	model (interface{}): 模型对象
//	conditions (...interface{}): 查询条件
//
// 返回:
//
//	bool: 是否存在
//	error: 错误信息
func (s *MySQLStorage) Exists(ctx context.Context, model interface{}, conditions ...interface{}) (bool, error) {
	s.logger.Debug("正在检查记录是否存在", interfaces.String("model", fmt.Sprintf("%T", model)))
	count, err := s.Count(ctx, model, conditions...)
	if err != nil {
		// Count 内部已记录日志
		return false, err
	}
	exists := count > 0
	s.logger.Debug("检查记录是否存在完成", interfaces.String("model", fmt.Sprintf("%T", model)), interfaces.Bool("exists", exists))
	return exists, nil
}

// Transaction 执行事务
//
// 描述: 在事务中执行操作
//
// 参数:
//
//	ctx (context.Context): 上下文
//	fn (func(ctx context.Context) error): 事务函数
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	s.logger.Debug("开始执行事务...")
	if s.db == nil {
		err := errors.New("MySQL连接未建立")
		s.logger.Error("执行事务失败，MySQL连接未建立", interfaces.ErrorField(err))
		return err
	}

	if fn == nil {
		return errors.New("事务函数不能为空")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建新的storage实例，使用事务连接
		txStorage := &MySQLStorage{
			db:     tx,
			config: s.config,
			logger: s.logger, // 传递logger
		}

		// 在上下文中传递事务storage
		txCtx := context.WithValue(ctx, "tx_storage", txStorage)
		return fn(txCtx)
	})

	if err != nil {
		s.logger.Error("事务执行失败，已回滚", interfaces.ErrorField(err))
	} else {
		s.logger.Debug("COMMIT TRANSACTION")
	}
	return err
}

// Raw 执行原生SQL
//
// 描述: 执行原生SQL语句
//
// 参数:
//
//	ctx (context.Context): 上下文
//	sql (string): SQL语句
//	values (...interface{}): 参数值
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Raw(ctx context.Context, sql string, values ...interface{}) error {
	s.logger.Debug("正在执行原生SQL", interfaces.String("sql", sql))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if sql == "" {
		return errors.New("SQL语句不能为空")
	}

	err := s.db.WithContext(ctx).Exec(sql, values...).Error
	if err != nil {
		wrappedErr := errors.Wrap(err, "执行原生SQL失败")
		s.logger.Error("执行原生SQL失败", interfaces.ErrorField(wrappedErr), interfaces.String("sql", sql))
		return wrappedErr
	}
	s.logger.Debug("原生SQL执行成功")
	return nil
}

// Migrate 迁移表结构
//
// 描述: 自动迁移数据库表结构
//
// 参数:
//
//	ctx (context.Context): 上下文
//	models (...interface{}): 模型对象
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) Migrate(ctx context.Context, models ...interface{}) error {
	s.logger.Debug(fmt.Sprintf("准备迁移 %d 个模型...", len(models)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if len(models) == 0 {
		return errors.New("迁移的模型不能为空")
	}

	err := s.db.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		wrappedErr := errors.Wrap(err, "迁移表结构失败")
		s.logger.Error("迁移表结构失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}
	s.logger.Debug("迁移表结构成功")
	return nil
}

// GetDB 获取GORM数据库实例
//
// 描述: 获取底层的GORM数据库实例
//
// 返回:
//
//	*gorm.DB: GORM数据库实例
func (s *MySQLStorage) GetDB() *gorm.DB {
	s.logger.Debug("获取底层GORM DB实例")
	return s.db
}

// GetConfig 获取配置
//
// 描述: 获取MySQL配置信息
//
// 返回:
//
//	*MySQLConfig: MySQL配置
func (s *MySQLStorage) GetConfig() *MySQLConfig {
	s.logger.Debug("获取MySQL配置")
	return s.config
}

// InitializeModels 初始化所有模型
//
// 函数名: InitializeModels
//
// 描述: 在应用启动时初始化所有数据库表结构，统一管理所有模型的迁移
//
// 参数:
//
//	ctx (context.Context): 上下文
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) InitializeModels(ctx context.Context) error {
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	s.logger.Info("开始初始化数据库表结构...")

	// 定义所有需要迁移的模型，按依赖关系排序
	// 基础模型
	basicModels := []interface{}{
		// &models.User{},
	}

	// 图片相关模型
	imageModels := []interface{}{}

	// 分批迁移模型
	modelGroups := []struct {
		name   string
		models []interface{}
	}{
		{"基础模型", basicModels},
		{"图片模型", imageModels},
	}

	for _, group := range modelGroups {
		s.logger.Info(fmt.Sprintf("正在迁移%s...", group.name))

		if err := s.db.WithContext(ctx).AutoMigrate(group.models...); err != nil {
			wrappedErr := errors.Wrapf(err, "迁移%s失败", group.name)
			s.logger.Error("迁移失败", interfaces.ErrorField(wrappedErr), interfaces.String("group", group.name))
			return wrappedErr
		}

		s.logger.Info(fmt.Sprintf("%s迁移完成", group.name))
	}

	s.logger.Info("所有数据库表初始化完成")
	return nil
}

// MigrateSpecificModels 迁移指定模型
//
// 函数名: MigrateSpecificModels
//
// 描述: 迁移指定的模型，用于按需迁移
//
// 参数:
//
//	ctx (context.Context): 上下文
//	models (...interface{}): 要迁移的模型
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) MigrateSpecificModels(ctx context.Context, models ...interface{}) error {
	s.logger.Debug(fmt.Sprintf("准备迁移 %d 个指定模型...", len(models)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if len(models) == 0 {
		return errors.New("迁移的模型不能为空")
	}

	err := s.db.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		wrappedErr := errors.Wrap(err, "迁移指定模型失败")
		s.logger.Error("迁移指定模型失败", interfaces.ErrorField(wrappedErr))
		return wrappedErr
	}
	s.logger.Debug("迁移指定模型成功")
	return nil
}

// DropTables 删除指定表
//
// 函数名: DropTables
//
// 描述: 删除指定的数据库表，用于开发和测试环境
//
// 参数:
//
//	ctx (context.Context): 上下文
//	models (...interface{}): 要删除的模型对应的表
//
// 返回:
//
//	error: 错误信息
func (s *MySQLStorage) DropTables(ctx context.Context, models ...interface{}) error {
	s.logger.Warn(fmt.Sprintf("准备删除 %d 个表...", len(models)))
	if s.db == nil {
		return errors.New("MySQL连接未建立")
	}

	if len(models) == 0 {
		return errors.New("要删除的表不能为空")
	}

	migrator := s.db.WithContext(ctx).Migrator()

	for _, model := range models {
		modelName := fmt.Sprintf("%T", model)
		s.logger.Debug(fmt.Sprintf("  - 正在检查表 %s", modelName))
		if migrator.HasTable(model) {
			s.logger.Debug(fmt.Sprintf("  - 正在删除表 %s", modelName))
			if err := migrator.DropTable(model); err != nil {
				wrappedErr := errors.Wrapf(err, "删除表失败: %s", modelName)
				s.logger.Error("删除表失败", interfaces.ErrorField(wrappedErr), interfaces.String("model", modelName))
				return wrappedErr
			}
		} else {
			s.logger.Debug(fmt.Sprintf("  - 表 %s 不存在，跳过删除", modelName))
		}
	}

	s.logger.Warn("所有指定表已删除")
	return nil
}

// CheckTableExists 检查表是否存在
//
// 函数名: CheckTableExists
//
// 描述: 检查指定模型对应的表是否存在
//
// 参数:
//
//	ctx (context.Context): 上下文
//	model (interface{}): 模型对象
//
// 返回:
//
//	bool: 表是否存在
//	error: 错误信息
func (s *MySQLStorage) CheckTableExists(ctx context.Context, model interface{}) (bool, error) {
	modelName := fmt.Sprintf("%T", model)
	s.logger.Debug("正在检查表是否存在", interfaces.String("model", modelName))
	if s.db == nil {
		return false, errors.New("MySQL连接未建立")
	}

	if model == nil {
		return false, errors.New("模型对象不能为空")
	}

	migrator := s.db.WithContext(ctx).Migrator()
	exists := migrator.HasTable(model)

	s.logger.Debug("检查表是否存在完成", interfaces.String("model", modelName), interfaces.Bool("exists", exists))
	return exists, nil
}
