package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator 验证器
type Validator struct {
	validate *validator.Validate
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// ValidateStruct 验证结构体
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsg string
		for _, e := range validationErrors {
			errMsg += fmt.Sprintf("%s: %s; ", e.Field(), getErrorMsg(e))
		}
		return fmt.Errorf(errMsg)
	}
	return err
}

// getErrorMsg 获取错误消息
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	default:
		return "is invalid"
	}
}
