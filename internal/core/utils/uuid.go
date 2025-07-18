package utils

import (
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// GenerateUUID 生成标准UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateShortUUID 生成短UUID（去掉连字符）
func GenerateShortUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// ValidateUUID 验证UUID格式
func ValidateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.Wrap(err, "invalid UUID format")
	}
	return nil
}

// IsValidUUID 检查是否为有效UUID
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
