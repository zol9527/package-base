package utils

import (
	"os"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var bcryptCost atomic.Int32

func init() {
	// 允许通过环境变量 BCRYPT_COST 动态设置，非法或缺失则使用默认值
	bcryptCost.Store(int32(bcrypt.DefaultCost))
	if v := os.Getenv("BCRYPT_COST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= bcrypt.MinCost && n <= bcrypt.MaxCost {
			bcryptCost.Store(int32(n))
		}
	}
}

// Function: SetBcryptCost
//
// Description: 运行期动态设置 bcrypt cost（测试或性能基准环境使用），非法值将被忽略。
//
// Parameters:
//
//	cost(int): 期望值 (bcrypt.MinCost ~ bcrypt.MaxCost)
//
// Returns:
//
//	ok(bool): 是否成功应用
func SetBcryptCost(cost int) bool {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return false
	}
	bcryptCost.Store(int32(cost))
	return true
}

// Function: GetBcryptCost
//
// Description: 获取当前生效的 bcrypt cost。
//
// Parameters:
//
// Returns:
//
//	cost(int): 当前 cost
func GetBcryptCost() int { return int(bcryptCost.Load()) }

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), GetBcryptCost())
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
