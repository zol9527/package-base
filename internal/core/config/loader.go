package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Loader 配置加载器
type Loader struct {
	config *Config
}

// NewLoader 创建配置加载器
func NewLoader() *Loader {
	return &Loader{
		config: DefaultConfig(),
	}
}

// LoadFromFile 从文件加载配置
func (l *Loader) LoadFromFile(configPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.Wrapf(err, "config file does not exist: %s", configPath)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}

	// 根据文件扩展名解析
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		return l.loadFromJSON(data)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// LoadFromEnv 从环境变量加载配置
func (l *Loader) LoadFromEnv() error {
	// 应用配置
	if name := os.Getenv("APP_NAME"); name != "" {
		l.config.App.Name = name
	}
	if version := os.Getenv("APP_VERSION"); version != "" {
		l.config.App.Version = version
	}
	if mode := os.Getenv("APP_MODE"); mode != "" {
		l.config.App.Mode = mode
	}
	if debug := os.Getenv("APP_DEBUG"); debug != "" {
		l.config.App.Debug = strings.ToLower(debug) == "true"
	}

	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err == nil {
			l.config.Server.Port = portInt
		}
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		l.config.Server.Host = host
	}

	// 数据库配置
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		l.config.Database.MySQL.Host = host
	}
	if port := os.Getenv("MYSQL_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err == nil {
			l.config.Database.MySQL.Port = portInt
		}
	}
	if username := os.Getenv("MYSQL_USERNAME"); username != "" {
		l.config.Database.MySQL.Username = username
	}
	if password := os.Getenv("MYSQL_PASSWORD"); password != "" {
		l.config.Database.MySQL.Password = password
	}
	if database := os.Getenv("MYSQL_DATABASE"); database != "" {
		l.config.Database.MySQL.Database = database
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		l.config.Database.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		var portInt int
		if _, err := fmt.Sscanf(port, "%d", &portInt); err == nil {
			l.config.Database.Redis.Port = portInt
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		l.config.Database.Redis.Password = password
	}
	if database := os.Getenv("REDIS_DATABASE"); database != "" {
		var dbInt int
		if _, err := fmt.Sscanf(database, "%d", &dbInt); err == nil {
			l.config.Database.Redis.Database = dbInt
		}
	}

	// JWT配置
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		l.config.JWT.Secret = secret
	}
	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		l.config.JWT.Issuer = issuer
	}

	return nil
}

// loadFromJSON 从JSON数据加载配置
func (l *Loader) loadFromJSON(data []byte) error {
	return json.Unmarshal(data, l.config)
}

// GetConfig 获取配置
func (l *Loader) GetConfig() *Config {
	return l.config
}

// SetConfig 设置配置
func (l *Loader) SetConfig(config *Config) {
	l.config = config
}

// SaveToFile 保存配置到文件
func (l *Loader) SaveToFile(configPath string) error {
	// 创建目录
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create config directory")
	}

	// 根据文件扩展名保存
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		return l.saveToJSON(configPath)
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}
}

// saveToJSON 保存为JSON格式
func (l *Loader) saveToJSON(configPath string) error {
	data, err := json.MarshalIndent(l.config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config to JSON")
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadConfig 便捷函数：加载配置
func LoadConfig(configPath string) (*Config, error) {
	loader := NewLoader()

	// 从环境变量加载基础配置
	if err := loader.LoadFromEnv(); err != nil {
		return nil, errors.Wrap(err, "failed to load config from environment")
	}

	// 如果指定了配置文件，则从文件加载
	if configPath != "" {
		if err := loader.LoadFromFile(configPath); err != nil {
			return nil, errors.Wrap(err, "failed to load config from file")
		}
	}

	config := loader.GetConfig()

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config validation failed")
	}

	return config, nil
}
