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

// OpenAIProvider OpenAI 提供商实现
type OpenAIProvider struct {
	client *openai.Client
	config *config.OpenAIConfig
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(cfg *config.OpenAIConfig) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API Key 未配置")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)

	// 设置自定义 BaseURL（如果提供）
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	// 设置组织 ID（如果提供）
	if cfg.OrgID != "" {
		clientConfig.OrgID = cfg.OrgID
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &OpenAIProvider{
		client: client,
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// Enabled 返回是否已正确配置
func (p *OpenAIProvider) Enabled() bool {
	return p.client != nil && p.config.APIKey != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *OpenAIProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	model := p.config.Model
	if model == "" {
		model = openai.GPT4Dot1Mini
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
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
		return "", "", fmt.Errorf("OpenAI API 调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf("OpenAI API 返回空结果")
	}

	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &out); err != nil {
		return "", "", fmt.Errorf("解析 OpenAI 响应失败: %w", err)
	}

	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}
