package llm

import "fmt"

// LLMError 定义 LLM 相关错误类型
type LLMError struct {
	Type    ErrorType
	Message string
	Err     error
}

// ErrorType 定义错误类型枚举
type ErrorType int

const (
	ErrorTypeAuth ErrorType = iota
	ErrorTypeTimeout
	ErrorTypeQuota
	ErrorTypeNetwork
	ErrorTypeGeneral
)

// Error 实现 error 接口
func (e *LLMError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 支持错误链
func (e *LLMError) Unwrap() error {
	return e.Err
}

// NewAuthError 创建认证错误
func NewAuthError(msg string, err error) *LLMError {
	return &LLMError{
		Type:    ErrorTypeAuth,
		Message: msg,
		Err:     err,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(msg string, err error) *LLMError {
	return &LLMError{
		Type:    ErrorTypeTimeout,
		Message: msg,
		Err:     err,
	}
}

// NewQuotaError 创建配额错误
func NewQuotaError(msg string, err error) *LLMError {
	return &LLMError{
		Type:    ErrorTypeQuota,
		Message: msg,
		Err:     err,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(msg string, err error) *LLMError {
	return &LLMError{
		Type:    ErrorTypeNetwork,
		Message: msg,
		Err:     err,
	}
}

// NewGeneralError 创建一般错误
func NewGeneralError(msg string, err error) *LLMError {
	return &LLMError{
		Type:    ErrorTypeGeneral,
		Message: msg,
		Err:     err,
	}
}