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
