package utils

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash password")
	}
	return string(hashedBytes), nil
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.Wrap(err, "password verification failed")
	}
	return nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 100 {
		return errors.New("password must be less than 100 characters long")
	}

	// 检查是否包含至少一个数字
	hasNumber := false
	// 检查是否包含至少一个字母
	hasLetter := false

	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasNumber = true
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasLetter {
		return errors.New("password must contain at least one letter")
	}

	return nil
}
