package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Validator 验证器
type Validator struct {
	validate *validator.Validate
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	validate := validator.New()

	// 注册自定义验证器
	validate.RegisterValidation("alphanum", validateAlphaNumeric)

	return &Validator{
		validate: validate,
	}
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(data interface{}) error {
	err := v.validate.Struct(data)
	if err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, v.formatValidationError(err))
		}
		return errors.New(strings.Join(validationErrors, "; "))
	}
	return nil
}

// formatValidationError 格式化验证错误信息
func (v *Validator) formatValidationError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be no more than %s characters long", field, err.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// validateAlphaNumeric 自定义字母数字验证器
func validateAlphaNumeric(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	validate := validator.New()
	if err := validate.Var(email, "required,email"); err != nil {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) error {
	validate := validator.New()
	validate.RegisterValidation("alphanum", validateAlphaNumeric)

	if err := validate.Var(username, "required,min=3,max=50,alphanum"); err != nil {
		if strings.Contains(err.Error(), "alphanum") {
			return errors.New("username must contain only alphanumeric characters")
		}
		if strings.Contains(err.Error(), "min") {
			return errors.New("username must be at least 3 characters long")
		}
		if strings.Contains(err.Error(), "max") {
			return errors.New("username must be no more than 50 characters long")
		}
		return errors.New("invalid username format")
	}
	return nil
}

// SanitizeInput 清理输入数据
func SanitizeInput(input string) string {
	// 移除前后空格
	cleaned := strings.TrimSpace(input)

	// 移除危险字符
	dangerousChars := []string{"<", ">", "&", "\"", "'", "/", "\\"}
	for _, char := range dangerousChars {
		cleaned = strings.ReplaceAll(cleaned, char, "")
	}

	return cleaned
}

// IsEmpty 检查值是否为空
func IsEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}
