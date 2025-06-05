package llm

import (
	"context"
	"fmt"

	"termi.sh/termi/internal/config"
	"termi.sh/termi/internal/llm/providers"
)

// Provider 定义 LLM 提供商接口
type Provider interface {
	// AskSmart 根据用户 query 返回 command 或 ask
	// 如果需要更多信息，则 ask 字段非空
	AskSmart(ctx context.Context, prompt string) (command string, ask string, err error)
	
	// Name 返回提供商名称
	Name() string
	
	// Enabled 返回是否已正确配置
	Enabled() bool
}

var currentProvider Provider

// Initialize 初始化 LLM 提供商
func Initialize(cfg *config.Config) error {
	var provider Provider
	var err error
	
	switch cfg.LLM.Provider {
	case config.ProviderOpenAI:
		if cfg.LLM.OpenAI == nil {
			return fmt.Errorf("OpenAI 配置未找到")
		}
		provider, err = providers.NewOpenAIProvider(cfg.LLM.OpenAI)
	case config.ProviderAzureOpenAI:
		if cfg.LLM.AzureOpenAI == nil {
			return fmt.Errorf("Azure OpenAI 配置未找到")
		}
		provider, err = providers.NewAzureOpenAIProvider(cfg.LLM.AzureOpenAI)
	case config.ProviderGemini:
		if cfg.LLM.Gemini == nil {
			return fmt.Errorf("Gemini 配置未找到")
		}
		provider, err = providers.NewGeminiProvider(cfg.LLM.Gemini)
	case config.ProviderClaude:
		if cfg.LLM.Claude == nil {
			return fmt.Errorf("Claude 配置未找到")
		}
		provider, err = providers.NewClaudeProvider(cfg.LLM.Claude)
	case config.ProviderLlamaCPP:
		if cfg.LLM.LlamaCPP == nil {
			return fmt.Errorf("Llama-cpp 配置未找到")
		}
		provider, err = providers.NewLlamaCPPProvider(cfg.LLM.LlamaCPP)
	default:
		return fmt.Errorf("不支持的 LLM 提供商: %s", cfg.LLM.Provider)
	}
	
	if err != nil {
		return fmt.Errorf("初始化 LLM 提供商失败: %w", err)
	}
	
	currentProvider = provider
	return nil
}

// Enabled 返回是否已正确配置 LLM
func Enabled() bool {
	return currentProvider != nil && currentProvider.Enabled()
}

// AskSmart 根据用户 query 返回 command 或 ask
// 如果需要更多信息，则 ask 字段非空
func AskSmart(prompt string) (command string, ask string, err error) {
	if currentProvider == nil {
		return "", "", fmt.Errorf("LLM 提供商未初始化")
	}
	
	if !currentProvider.Enabled() {
		return "", "", fmt.Errorf("LLM 提供商 %s 未正确配置", currentProvider.Name())
	}
	
	ctx := context.Background()
	return currentProvider.AskSmart(ctx, prompt)
}

// GetProviderName 返回当前提供商名称
func GetProviderName() string {
	if currentProvider == nil {
		return "未知"
	}
	return currentProvider.Name()
}
