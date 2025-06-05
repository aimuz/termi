package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"termi.sh/termi/internal/config"
)

// ClaudeProvider Claude 提供商实现
type ClaudeProvider struct {
	client *anthropic.Client
	config *config.ClaudeConfig
}

// NewClaudeProvider 创建 Claude 提供商
func NewClaudeProvider(cfg *config.ClaudeConfig) (*ClaudeProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Claude API Key 未配置")
	}
	
	options := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	
	// 设置自定义 BaseURL（如果提供）
	if cfg.BaseURL != "" {
		options = append(options, option.WithBaseURL(cfg.BaseURL))
	}
	
	client := anthropic.NewClient(options...)
	
	return &ClaudeProvider{
		client: &client,
		config: cfg,
	}, nil
}

// Name 返回提供商名称
func (p *ClaudeProvider) Name() string {
	return "Claude"
}

// Enabled 返回是否已正确配置
func (p *ClaudeProvider) Enabled() bool {
	return p.client != nil && p.config.APIKey != ""
}

// AskSmart 根据用户 query 返回 command 或 ask
func (p *ClaudeProvider) AskSmart(ctx context.Context, prompt string) (command string, ask string, err error) {
	timeout := time.Duration(p.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	model := p.config.Model
	if model == "" {
		model = "claude-3-haiku-20240307"
	}
	
	systemPrompt := `你是 Linux 命令行专家。根据用户需求和对话历史，生成合适的 Bash 命令。

如果信息充足，返回 JSON {"command":"..."}，其中 command 是可直接执行的 Bash 命令。
如果需要更多信息，返回 JSON {"ask":"..."}，ask 用中文向用户提出具体的补充问题。

注意：
- 仔细理解用户的完整意图和上下文
- 如果之前的对话中已经提供了相关信息，请充分利用
- 生成的命令应该是安全、准确且可执行的`
	
	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(1000),
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Temperature: anthropic.Float(0.2),
	})
	if err != nil {
		return "", "", fmt.Errorf("Claude API 调用失败: %w", err)
	}
	
	if len(message.Content) == 0 {
		return "", "", fmt.Errorf("Claude API 返回空结果")
	}
	
	// 提取响应文本
	var responseText string
	for _, content := range message.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}
	
	if responseText == "" {
		return "", "", fmt.Errorf("Claude API 返回空文本")
	}
	
	// 解析 JSON 响应
	var out struct {
		Command string `json:"command"`
		Ask     string `json:"ask"`
	}
	if err := json.Unmarshal([]byte(responseText), &out); err != nil {
		return "", "", fmt.Errorf("解析 Claude 响应失败: %w, 原始响应: %s", err, responseText)
	}
	
	return strings.TrimSpace(out.Command), strings.TrimSpace(out.Ask), nil
}