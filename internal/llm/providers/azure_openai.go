package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"termi.sh/termi/internal/config"
)

// AzureOpenAIProvider Azure OpenAI 提供商实现
type AzureOpenAIProvider struct {
	client *openai.Client
	config *config.AzureOpenAIConfig
}

// NewAzureOpenAIProvider 创建 Azure OpenAI 提供商
func NewAzureOpenAIProvider(cfg *config.AzureOpenAIConfig) (*AzureOpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Azure OpenAI API Key 未配置")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("Azure OpenAI Base URL 未配置")
	}
	if cfg.DeploymentID == "" {
		return nil, fmt.Errorf("Azure OpenAI Deployment ID 未配置")
	}

	clientConfig := openai.DefaultAzureConfig(cfg.APIKey, cfg.BaseURL)
	clientConfig.APIVersion = cfg.APIVersion
	if clientConfig.APIVersion == "" {
		clientConfig.APIVersion = "2023-12-01-preview"
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &AzureOpenAIProvider{
		client: client,
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *AzureOpenAIProvider) Name() string {
	return "Azure OpenAI"
}

// Enabled 返回是否已正确配置
func (p *AzureOpenAIProvider) Enabled() bool {
	return p.client != nil && p.config.APIKey != "" && p.config.BaseURL != "" && p.config.DeploymentID != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *AzureOpenAIProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.config.DeploymentID, // Azure 使用 deployment ID 作为模型名
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt(),
			},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature:    0.2,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
	if err != nil {
		return "", "", fmt.Errorf("Azure OpenAI API 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf("Azure OpenAI API 返回空结果")
	}

	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return "", "", fmt.Errorf("解析 Azure OpenAI 响应失败: %w", err)
	}

	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}
